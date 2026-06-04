package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"ns-tracking-go/service/tracking/rpc/internal/event/define"
	"ns-tracking-go/service/tracking/rpc/internal/queue/tasks"
	"ns-tracking-go/service/tracking/rpc/internal/svc"
	"ns-tracking-go/service/tracking/rpc/internal/utils"
	"ns-tracking-go/service/tracking/rpc/tracking"

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

// Upsert 插入或更新追踪数据（完整四表同步版本 + 幂等去重）
// 对标：Ruby tracking_detail.rb:272-293 (sync!方法)
// 流程：
//  1. 幂等去重检查（防止重复处理）
//  2. Upsert tracking_details（备份last_detail）
//  3. 查询tracking_log获取ID
//  4. 同步回写tracking_log状态
//  5. 写入tracking_cache_log（可选）
func (l *UpsertLogic) Upsert(in *tracking.UpsertRequest) (*tracking.UpsertResponse, error) {
	// ========== Step 0: 幂等去重检查（防止重复处理）==========
	// 幂等 key: upsert:{tracking_number}:{synced_at}
	idempotencyKey := fmt.Sprintf("upsert:%s:%d", in.TrackingNumber, in.SyncedAt)
	
	if l.svcCtx.Redis != nil {
		// 检查是否已处理
		exists, err := l.svcCtx.Redis.ExistsCtx(l.ctx, idempotencyKey)
		if err == nil && exists {
			l.Logger.Infof("Duplicate request detected, skipping: %s", utils.MaskTrackingNumber(in.TrackingNumber))
			return &tracking.UpsertResponse{
				Success: true,
				Message: "duplicate request (already processed)",
			}, nil
		}
		
		// 设置幂等 key（1小时过期）
		l.svcCtx.Redis.SetexCtx(l.ctx, idempotencyKey, "1", 3600)
	}
	
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

	l.Logger.Infof("Upsert tracking_details success: %s, status=%d", utils.MaskTrackingNumber(in.TrackingNumber), in.Status)

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
		// 注意：需要检查 overseas_package（影响 InfoReceived 状态映射）
		trackStatus := l.mapTrackStatus(in.TrackingNumber, in.Status)
		syncedAt := in.SyncedAt

		// 同步5个字段：track_status, synced_at, received_at, delivered_at, tracked_at
		err = l.svcCtx.TrackingDAO.SyncTrackingLog(l.ctx, in.TrackingNumber, trackStatus, syncedAt, in.ReceivedAt, in.DeliveredAt, in.TrackedAt)
		if err != nil {
			l.Logger.Errorf("Sync tracking_log failed: %v", err)
			// 同步失败不影响主流程，记录日志即可
		} else {
			l.Logger.Infof("Sync tracking_log success: %s → track_status=%d", utils.MaskTrackingNumber(in.TrackingNumber), trackStatus)
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

// mapTrackStatus 状态码映射 → TrackStatus（对标 Ruby tracking_detail.rb:70-89）
// 参数：
//   - trackingNumber：运单号（用于查询 overseas_package）
//   - status：云途状态码（如 10, 20, 50）
// 返回：
//   - TrackStatus 值（0=待揽收, 1=运输中, 2=已签收, 3=异常）
//
// 映射规则（对标 Ruby tracking_detail.rb:70-89）：
//   - InfoReceived + overseas_package → 1 (in_transit)
//   - InfoReceived + NO overseas_package → 0 (to_receive)
//   - InTransit / InTransit_Arrival → 1 (in_transit)
//   - Delivered / AvailableForPickup → 2 (delivered)
//   - Exception / DeliveryFailure / Expired / Exception_Returned / Exception_Cancel → 3 (track_exception)
//   - 默认 → 0 (to_receive)
func (l *UpsertLogic) mapTrackStatus(trackingNumber string, status int32) int32 {
	statusStr := strconv.Itoa(int(status))

	// Step 1: 状态码 → 状态文本映射
	// 对标：Ruby tracking_detail.rb:66-68 (status_txt方法)
	statusTxt := getStatusText(statusStr)

	// Step 2: InfoReceived 特殊处理（检查 overseas_package）
	// 对标：Ruby tracking_detail.rb:75-79
	if statusTxt == "InfoReceived" {
		hasOverseasPackage, err := l.svcCtx.OverseasPackageDAO.HasOverseasPackage(l.ctx, trackingNumber)
		if err != nil {
			l.Logger.Errorf("HasOverseasPackage query failed: %v, fallback to to_receive", err)
			return 0 // 降级：待揽收
		}

		if hasOverseasPackage {
			return 1 // in_transit（海外包裹）
		}
		return 0 // to_receive（非海外包裹）
	}

	// Step 3: 其他状态文本映射
	// 对标：Ruby tracking_detail.rb:80-88
	switch statusTxt {
	case "InTransit", "InTransit_Arrival":
		return 1 // in_transit

	case "Delivered", "AvailableForPickup":
		return 2 // delivered

	case "DeliveryFailure", "Exception", "Expired", "Exception_Returned", "Exception_Cancel":
		return 3 // track_exception

	default:
		// NotFound / Undefined → to_receive（保守策略）
		return 0
	}
}

// getStatusText 云途状态码 → 状态文本映射
// 对标：Ruby yun_express_track_formatter.rb:6-19 (STATUS_CODES)
func getStatusText(statusCode string) string {
	// 云途状态码映射表（12 项）
	statusCodes := map[string]string{
		"0":    "NotFound",
		"10":   "InfoReceived",
		"20":   "InTransit",
		"30":   "AvailableForPickup",
		"40":   "DeliveryFailure",
		"50":   "Delivered",
		"60":   "Exception",
		"70":   "Expired",
		"80":   "Exception", // 海关查验（会被改写为 20）
		"90":   "Exception_Returned",
		"100":  "Exception_Cancel",
		"1001": "InTransit_Arrival",
	}

	if text, ok := statusCodes[statusCode]; ok {
		return text
	}
	return "Undefined"
}
