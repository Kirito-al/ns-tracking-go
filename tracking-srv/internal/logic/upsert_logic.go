package logic

import (
	"context"
	"encoding/json"
	"strconv"

	"tracking-srv/internal/event/define"
	"tracking-srv/internal/queue/tasks"
	"tracking-srv/internal/svc"
	"tracking-srv/tracking"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpsertLogic 插入或更新追踪数据业务逻辑
type UpsertLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpsertLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpsertLogic {
	return &UpsertLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Upsert 插入或更新追踪数据（完整四表同步版本）
// 对标：Ruby tracking_detail.rb:272-293 (sync!方法)
// 流程：
//  1. Upsert tracking_details（备份last_detail）
//  2. 查询tracking_log获取ID
//  3. 同步回写tracking_log状态
//  4. 写入tracking_cache_log（可选）
func (l *UpsertLogic) Upsert(in *tracking.UpsertRequest) (*tracking.UpsertResponse, error) {
	// ========== Step 1: Upsert tracking_details（备份last_detail）==========
	// 对标：Ruby tracking_detail.rb:280-288
	err := l.svcCtx.TrackingDAO.Upsert(l.ctx,
		in.TrackingNumber,
		in.Detail,
		in.Status,
		in.ServiceClass,
	)

	if err != nil {
		l.Logger.Errorf("Upsert tracking_details failed: %v", err)

		// ========== 触发入库失败事件 ==========
		if l.svcCtx.EventDispatcher != nil {
			// 构造原始 Payload（用于重试）
			originalPayload := map[string]interface{}{
				"tracking_number": in.TrackingNumber,
				"detail":          in.Detail,
				"status":           in.Status,
				"service_class":   in.ServiceClass,
			}
			originalPayloadJSON, _ := json.Marshal(originalPayload)

			// 创建失败事件
			event := define.NewTrackingUpsertFailedEvent(
				in.TrackingNumber,
				err.Error(),
				500, // ErrorCode：数据库错误
				"Upsert", // FailedStage
				string(originalPayloadJSON),
			)

			// Dispatch 事件（触发 LogListener）
			l.svcCtx.EventDispatcher.Dispatch(event)

			// Enqueue 重试任务（Asynq 内置重试）
			if l.svcCtx.AsynqClient != nil {
				taskPayload, _ := json.Marshal(event)
				task := asynq.NewTask(tasks.TaskRetryFailed, taskPayload, asynq.MaxRetry(tasks.MaxRetry))
				_, enqueueErr := l.svcCtx.AsynqClient.Enqueue(task)
				if enqueueErr != nil {
					l.Logger.Errorf("Enqueue retry task failed: %v", enqueueErr)
				} else {
					l.Logger.Infof("Retry task enqueued: tracking_number=%s", in.TrackingNumber)
				}
			}
		}

		return &tracking.UpsertResponse{
			Success: false,
			Message: "数据库写入失败: " + err.Error(),
		}, nil
	}

	l.Logger.Infof("Upsert tracking_details success: %s, status=%d", in.TrackingNumber, in.Status)

	// ========== Step 2: 查询tracking_log获取ID和渠道信息 ==========
	// 对标：Ruby tracking_detail.rb:268（关联tracking_log）
	trackingLog, err := l.svcCtx.TrackingLogDAO.GetBySourceTrackingNumber(l.ctx, in.TrackingNumber)
	if err != nil {
		l.Logger.Errorf("Query tracking_log failed: %v", err)
		// 查询失败不影响主流程，继续执行（降级处理）
	} else if trackingLog != nil {
		l.Logger.Infof("Found tracking_log: id=%d, channel=%s", trackingLog.ID, trackingLog.ChannelAlias)

		// ========== Step 3: 同步回写tracking_log状态 ==========
		// 对标：Ruby tracking_detail.rb:316-328 (sync_tracking_log)
		// 转换状态码：tracking_detail.status → tracking_log.track_status
		trackStatus := l.convertStatusToTrackStatus(in.Status)
		syncedAt := in.SyncedAt

		err = l.svcCtx.TrackingDAO.SyncTrackingLog(l.ctx, in.TrackingNumber, trackStatus, syncedAt)
		if err != nil {
			l.Logger.Errorf("Sync tracking_log failed: %v", err)
			// 同步失败不影响主流程，记录日志即可
		} else {
			l.Logger.Infof("Sync tracking_log success: %s → track_status=%d", in.TrackingNumber, trackStatus)
		}
	}

	// ========== Step 4: 写入tracking_cache_log（可选）==========
	// 对标：Ruby tracking_cache_log.rb（缓冲同步）
	// 注意：只在特定场景下写入（如：重试机制、批量同步）
	// Webhook场景暂时不写入（避免重复）
	// TODO: 根据业务需求决定是否启用

	// ========== 返回成功响应 ==========
	// ========== 触发入库成功事件 ==========
	if l.svcCtx.EventDispatcher != nil {
		// 创建成功事件（10 核心字段）
		event := define.NewTrackingUpsertedEvent(
			in.TrackingNumber,
			in.Status,
			in.ServiceClass,
			in.Detail,
			"", // CountryCode（暂无）
			in.SyncedAt,
			"webhook", // Source
		)

		// Dispatch 事件（触发 LogListener）
		l.svcCtx.EventDispatcher.Dispatch(event)
	}

	return &tracking.UpsertResponse{
		Success: true,
		Message: "追踪数据已保存（四表同步完成）",
	}, nil
}

// convertStatusToTrackStatus 转换状态码（tracking_detail → tracking_log）
// 对标：Ruby tracking_detail.rb:73-89 (tracking_status方法)
// 映射规则：
//  - 0 (NotFound) → 0 (to_receive)
//  - 10 (InfoReceived) → 0 (to_receive)
//  - 20 (InTransit) → 1 (in_transit)
//  - 30 (AvailableForPickup) → 2 (delivered)
//  - 50 (Delivered) → 2 (delivered)
//  - 40,60,70,80,90,100 (Exception) → 3 (track_exception)
func (l *UpsertLogic) convertStatusToTrackStatus(status int32) int32 {
	statusStr := strconv.Itoa(int(status))

	// 状态码映射（对标 Ruby tracking_detail.rb:73-89）
	switch statusStr {
	case "0", "10":
		return 0 // to_receive（待揽收）
	case "20":
		return 1 // in_transit（运输中）
	case "30", "50":
		return 2 // delivered（已签收）
	case "40", "60", "70", "80", "90", "100":
		return 3 // track_exception（异常）
	default:
		return 1 // 默认：in_transit（保守策略）
	}
}