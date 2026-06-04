# ns-tracking-go 架构评估报告

> 三维度评估：可维护性、可观测性、高可用性
> 评估时间：2026-06-03
> 项目地址：D:\gohome\demo1-gozero\ns-tracking-go

---

## 一、可维护性评估

### 1.1 当前架构优点

| 维度 | 评分 | 说明 |
|------|------|------|
| **代码分层清晰** | ⭐⭐⭐⭐⭐ | handler → logic → dao → model，标准四层架构 |
| **依赖注入设计** | ⭐⭐⭐⭐⭐ | ServiceContext 统一管理依赖，易于测试和替换 |
| **配置管理规范** | ⭐⭐⭐⭐ | YAML 配置 + 环境变量，支持多环境部署 |
| **目录命名规范** | ⭐⭐⭐⭐⭐ | internal/ 包隔离，防止外部依赖污染 |

---

### 1.2 存在问题（需优化）

#### 问题 1：缺少聚合层（Service Layer）

**现状**：
```
tracking-api/handler → tracking-api/logic → tracking-srv (gRPC)
                                              ↓
                                        tracking-srv/logic → tracking-srv/dao
```

**问题**：
- logic 层直接调用 dao，业务逻辑和数据访问混杂
- 复杂业务场景（如：多表关联查询）难以复用

**优化建议**：新增 Service 层

```
tracking-srv/logic → tracking-srv/service → tracking-srv/dao
```

**新增文件**：
```
tracking-srv/internal/service/
├── tracking_service.go        # 轨迹业务服务（聚合 DAO）
├── token_service.go           # Token 路由服务
└── formatter_service.go       # 格式化服务
```

**示例代码**：
```go
// tracking-srv/internal/service/tracking_service.go
type TrackingService struct {
    trackingDAO    *dao.TrackingDAO
    trackingLogDAO *dao.TrackingLogDAO
    cache          *cache.TrackingCache
}

// UpsertWithCache：入库 + 缓存更新（聚合多个 DAO）
func (s *TrackingService) UpsertWithCache(ctx context.Context, ...) error {
    // 1. Upsert tracking_details
    s.trackingDAO.Upsert(...)
    
    // 2. Sync tracking_logs
    s.trackingLogDAO.Sync(...)
    
    // 3. Update cache
    s.cache.Set(...)
}
```

---

#### 问题 2：缺少 DTO 层（数据传输对象）

**现状**：
- Handler 直接返回 DB Model（`TrackingDetail`）
- Model 结构体包含 GORM tag，直接暴露给前端不安全

**优化建议**：新增 DTO 层

```
tracking-srv/internal/dto/
├── tracking_detail_dto.go      # 轨迹详情 DTO（前端返回）
├── webhook_payload_dto.go      # Webhook 接收 DTO
└── tracking_event_dto.go       # 轨迹事件 DTO
```

**示例代码**：
```go
// tracking-srv/internal/dto/tracking_detail_dto.go
type TrackingDetailDTO struct {
    TrackingNumber string         `json:"trackingNumber"`
    Status         int32          `json:"status"`
    StatusText     string         `json:"statusText"`
    Detail         json.RawMessage `json:"detail"`
    SyncedAt       int64          `json:"syncedAt"`
}

// ToDTO：Model → DTO 转换
func ToDTO(model *model.TrackingDetail) *TrackingDetailDTO {
    return &TrackingDetailDTO{
        TrackingNumber: model.TrackingNumber,
        Status:         model.Status,
        StatusText:     StatusCodes[model.Status],
        Detail:         json.RawMessage(model.Detail),
        SyncedAt:       model.SyncedAt,
    }
}
```

---

#### 问题 3：缺少常量层（Constants Layer）

**现状**：
- 状态码、错误码散落在各文件中
- 状态码映射表缺失（对标 Ruby STATUS_CODES）

**优化建议**：新增常量层

```
tracking-srv/internal/constants/
├── status_codes.go             # 状态码常量（对标 Ruby）
├── error_codes.go              # 错误码定义
├── channel_prefix.go           # 渠道前缀（GE-云途）
└── service_class.go            # 服务类名常量
```

**示例代码**：
```go
// tracking-srv/internal/constants/status_codes.go
const (
    StatusNotFound         = 0
    StatusInfoReceived     = 10
    StatusInTransit        = 20
    StatusAvailableForPickup = 30
    StatusDeliveryFailure  = 40
    StatusDelivered        = 50
    StatusException        = 60
    StatusExpired          = 70
    StatusExceptionReturned = 90
    StatusExceptionCancel  = 100
)

var StatusCodesMap = map[int32]string{
    0:   "NotFound",
    10:  "InfoReceived",
    20:  "InTransit",
    // ...对标 Ruby
}
```

---

#### 问题 4：缺少接口层（Interface Abstraction）

**现状**：
- DAO、Service 都是具体实现，无接口定义
- Mock 测试困难，替换实现需要修改大量代码

**优化建议**：新增接口定义

```
tracking-srv/internal/interfaces/
├── tracking_dao_interface.go   # DAO 接口定义
├── cache_interface.go          # 缓存接口
└── dispatcher_interface.go     # 事件分发接口（已存在）
```

**示例代码**：
```go
// tracking-srv/internal/interfaces/tracking_dao_interface.go
type ITrackingDAO interface {
    Upsert(ctx context.Context, trackingNumber, detail string, status int32) error
    GetLatest(ctx context.Context, trackingNumber string) (*model.TrackingDetail, error)
    MarkError(ctx context.Context, trackingNumber, errorMessage string) error
}

// Mock 实现（测试用）
type MockTrackingDAO struct{}

func (m *MockTrackingDAO) Upsert(...) error {
    return nil // Mock 逻辑
}
```

---

### 1.3 可维护性优化总表

| 问题 | 等级 | 优化建议 | 预计工作量 |
|------|------|---------|-----------|
| 缺少 Service 层 | 中 | 新增 service 目录，聚合 DAO | 2 小时 |
| 缺少 DTO 层 | 中 | 新增 dto 目录，Model → DTO 转换 | 1 小时 |
| 缺少 Constants 层 | 高 | 新增 constants 目录，对标 Ruby | 1 小时 |
| 缺少 Interface 层 | 低 | 新增 interfaces 目录（可选） | 1 小时 |

---

## 二、可观测性评估

### 2.1 当前架构缺失

| 维度 | 评分 | 说明 |
|------|------|------|
| **日志记录** | ⭐⭐⭐ | zap 终端日志，缺少结构化字段 |
| **监控指标** | ⭐ | 无 Prometheus 指标暴露 |
| **链路追踪** | ⭐ | 无 OpenTelemetry 集成 |
| **告警机制** | ⭐ | 无告警配置 |
| **健康检查** | ⭐⭐⭐ | 有 /health 端点，但缺少依赖检查 |

---

### 2.2 可观测性优化建议

#### 建议 1：结构化日志升级

**现状**：
```go
logx.Infof("Upsert success: %s, status=%d", trackingNumber, status)
```

**问题**：
- 日志缺少关键字段（TraceID、ServiceName、Env）
- 无法按字段聚合查询（如：Kibana/ElasticSearch）

**优化建议**：结构化日志（JSON 格式）

```go
// tracking-srv/internal/logger/structured_logger.go
type StructuredLogger struct {
    serviceName string
    env         string
}

func (l *StructuredLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
    traceID := ctx.Value("traceID")
    
    logEntry := map[string]interface{}{
        "timestamp":    time.Now().Unix(),
        "service":      l.serviceName,
        "env":          l.env,
        "trace_id":     traceID,
        "message":      msg,
        "tracking_number": fields["tracking_number"],
        "status":       fields["status"],
        "duration_ms":  fields["duration_ms"],
    }
    
    logx.InfoJson(logEntry) // JSON 格式输出
}
```

**日志输出示例**：
```json
{
  "timestamp": 1717405200,
  "service": "tracking-srv",
  "env": "production",
  "trace_id": "abc123",
  "message": "Upsert success",
  "tracking_number": "YT2606500704802225",
  "status": 20,
  "duration_ms": 45
}
```

---

#### 建议 2：Prometheus 指标暴露

**现状**：无监控指标

**优化建议**：新增 Prometheus metrics

```
tracking-srv/internal/metrics/
├── prometheus_metrics.go       # Prometheus 指标定义
├── metrics_middleware.go       # 指标收集中间件
```

**关键指标**：

| 指标类型 | 指标名称 | 说明 |
|----------|---------|------|
| Counter | `tracking_upsert_total` | 入库总次数 |
| Counter | `tracking_upsert_success` | 入库成功次数 |
| Counter | `tracking_upsert_failed` | 入库失败次数 |
| Histogram | `tracking_upsert_duration_ms` | 入库耗时分布 |
| Gauge | `tracking_cache_hit_rate` | 缓存命中率 |
| Gauge | `tracking_active_connections` | DB 连接数 |

**示例代码**：
```go
// tracking-srv/internal/metrics/prometheus_metrics.go
var (
    UpsertTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tracking_upsert_total",
            Help: "Total number of tracking upsert operations",
        },
        []string{"status", "service_class"},
    )
    
    UpsertDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "tracking_upsert_duration_ms",
            Help:    "Duration of tracking upsert operations",
            Buckets: []float64{10, 50, 100, 500, 1000},
        },
        []string{"service_class"},
    )
)

// 在 logic 中记录指标
func (l *UpsertLogic) Upsert(...) {
    start := time.Now()
    err := l.svcCtx.TrackingDAO.Upsert(...)
    
    duration := time.Since(start).Milliseconds()
    UpsertDuration.WithLabelValues("yunexpress").Observe(float64(duration))
    
    if err != nil {
        UpsertTotal.WithLabelValues("failed", "yunexpress").Inc()
    } else {
        UpsertTotal.WithLabelValues("success", "yunexpress").Inc()
    }
}
```

---

#### 建议 3：OpenTelemetry 集成（链路追踪）

**现状**：无链路追踪

**优化建议**：集成 OpenTelemetry

```
tracking-srv/internal/tracing/
├── otel_tracer.go              # OpenTelemetry Tracer 初始化
├── tracing_middleware.go       # 自动注入 TraceID
```

**链路流程**：
```
HTTP Request (tracking-api)
    ↓ TraceID: abc123
Webhook Handler
    ↓ Span: webhook.receive
gRPC Call (tracking-srv)
    ↓ Span: grpc.upsert
Upsert Logic
    ↓ Span: dao.upsert
    ↓ Span: event.dispatch
    ↓ Span: asynq.enqueue
```

**示例代码**：
```go
// tracking-srv/internal/tracing/otel_tracer.go
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptrace.New(context.Background(),
        otlptracehttp.NewClient())
    if err != nil {
        return nil, err
    }
    
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
        sdktrace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        )),
    )
    
    otel.SetTracerProvider(tp)
    return tp, nil
}

// 在 logic 中注入 TraceID
func (l *UpsertLogic) Upsert(ctx context.Context, ...) {
    tracer := otel.Tracer("tracking-srv")
    ctx, span := tracer.Start(ctx, "Upsert")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("tracking_number", in.TrackingNumber),
        attribute.Int64("status", in.Status),
    )
    
    // 业务逻辑...
}
```

---

#### 建议 4：健康检查增强

**现状**：
```go
// tracking-api/internal/handler/health/health_handler.go
func HealthHandler(ctx context.Context) {
    return {"status": "ok"}  // 仅返回 ok
}
```

**问题**：
- 不检查 DB 连接状态
- 不检查 Redis 连接状态
- 不检查 gRPC 服务状态

**优化建议**：依赖健康检查

```go
// tracking-api/internal/handler/health/health_handler.go
type HealthResponse struct {
    Status    string            `json:"status"`
    Checks    map[string]string `json:"checks"`
    Timestamp int64             `json:"timestamp"`
}

func HealthHandler(svcCtx *svc.ServiceContext) HealthResponse {
    checks := map[string]string{}
    
    // 1. 检查 DB
    if err := svcCtx.DB.Ping(); err != nil {
        checks["database"] = "unhealthy: " + err.Error()
    } else {
        checks["database"] = "healthy"
    }
    
    // 2. 检查 Redis
    if svcCtx.Redis != nil {
        if err := svcCtx.Redis.Ping(); err != nil {
            checks["redis"] = "unhealthy: " + err.Error()
        } else {
            checks["redis"] = "healthy"
        }
    }
    
    // 3. 检查 gRPC
    // ...
    
    overall := "healthy"
    for _, v := range checks {
        if strings.Contains(v, "unhealthy") {
            overall = "unhealthy"
        }
    }
    
    return HealthResponse{
        Status:    overall,
        Checks:    checks,
        Timestamp: time.Now().Unix(),
    }
}
```

---

#### 建议 5：告警机制配置

**优化建议**：新增告警配置

```
tracking-srv/config/alerts/
├── alert_rules.yml             # Prometheus 告警规则
└── alertmanager.yml            # AlertManager 配置
```

**告警规则示例**：
```yaml
# alert_rules.yml
groups:
  - name: tracking_alerts
    rules:
      - alert: HighUpsertFailureRate
        expr: rate(tracking_upsert_failed[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High upsert failure rate"
          description: "Upsert failure rate > 10% for 5 minutes"
      
      - alert: SlowUpsertDuration
        expr: histogram_quantile(0.99, tracking_upsert_duration_ms) > 500
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow upsert duration (P99 > 500ms)"
      
      - alert: LowCacheHitRate
        expr: tracking_cache_hit_rate < 0.7
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit rate (<70%)"
```

---

### 2.3 可观测性优化总表

| 建议 | 等级 | 实施步骤 | 预计工作量 |
|------|------|---------|-----------|
| 结构化日志 | 高 | JSON 格式 + TraceID | 1 小时 |
| Prometheus 指标 | 高 | metrics 目录 + middleware | 2 小时 |
| OpenTelemetry 集成 | 中 | tracing 目录 + Span 注入 | 3 小时 |
| 健康检查增强 | 高 | 依赖检查（DB/Redis/gRPC） | 1 小时 |
| 告警机制 | 中 | alert_rules.yml 配置 | 1 小时 |

---

## 三、高可用性评估

### 3.1 当前架构评估

| 维度 | 评分 | 说明 |
|------|------|------|
| **服务隔离** | ⭐⭐⭐⭐⭐ | tracking-api + tracking-srv 双服务架构 |
| **数据库连接池** | ⭐⭐⭐⭐⭐ | 已配置连接池（MaxOpenConns=100） |
| **失败重试机制** | ⭐⭐⭐⭐ | Asynq 内置重试（MaxRetry=3） |
| **降级逻辑** | ⭐ | 无降级逻辑（缓存 miss 无兜底） |
| **限流熔断** | ⭐ | 无限流配置 |
| **容灾备份** | ⭐ | 无异地备份 |

---

### 3.2 高可用性优化建议

#### 建议 1：降级逻辑实现

**现状**：
- Webhook 接收失败 → 直接返回 200（但数据丢失）
- DB 连接失败 → 服务直接 panic

**优化建议**：降级兜底策略

```go
// tracking-srv/internal/logic/upsert_logic.go
func (l *UpsertLogic) Upsert(ctx context.Context, in *tracking.UpsertRequest) {
    // 1. 尝试入库
    err := l.svcCtx.TrackingDAO.Upsert(...)
    
    if err != nil {
        // 2. 降级策略 A：写入 Redis 缓存（临时存储）
        if l.svcCtx.Redis != nil {
            cacheKey := "tracking:fallback:" + in.TrackingNumber
            l.svcCtx.Redis.Setex(cacheKey, in.Detail, 3600) // 1 小时
            logx.Errorf("Upsert failed, fallback to Redis cache: %s", in.TrackingNumber)
        }
        
        // 3. 降级策略 B：Enqueue Asynq 重试
        if l.svcCtx.AsynqClient != nil {
            task := asynq.NewTask("tracking:retry", payload)
            l.svcCtx.AsynqClient.Enqueue(task)
        }
        
        // 4. 仍然返回 200（云途不重试）
        return &tracking.UpsertResponse{
            Success: false,
            Message: "Upsert failed, fallback enabled",
        }, nil
    }
    
    return &tracking.UpsertResponse{Success: true}, nil
}
```

---

#### 建议 2：限流熔断配置

**现状**：无限流保护

**优化建议**：集成 Sentinel（阿里开源限流组件）

```go
// tracking-api/internal/middleware/limiter_middleware.go
import "github.com/alibaba/sentinel-golang/api"

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 限流规则：每秒最多 100 次 Webhook 接收
        if r.URL.Path == "/webhook/yunexpress/tracking/normal" {
            entry, blockError := api.Entry("webhook_receive")
            if blockError != nil {
                w.WriteHeader(http.StatusTooManyRequests)
                return
            }
            entry.Exit()
        }
        
        next.ServeHTTP(w, r)
    })
}
```

**限流规则配置**：
```yaml
# sentinel.yml
flowRules:
  - resource: webhook_receive
    limitApp: default
    grade: QPS
    count: 100  # 每秒 100 次
    strategy: Direct
```

---

#### 建议 3：数据库故障恢复

**现状**：
- DB 连接失败 → 服务 panic
- 无法自动恢复

**优化建议**：数据库重连机制

```go
// tracking-srv/internal/db/reconnect.go
type DBReconnector struct {
    db       *gorm.DB
    dsn      string
    maxRetry int
}

func (r *DBReconnector) PingAndReconnect() error {
    for i := 0; i < r.maxRetry; i++ {
        if err := r.db.Ping(); err != nil {
            logx.Errorf("DB ping failed (attempt %d): %v", i+1, err)
            
            // 重新连接
            newDB, err := gorm.Open(postgres.Open(r.dsn), &gorm.Config{})
            if err == nil {
                r.db = newDB
                logx.Info("DB reconnected successfully")
                return nil
            }
            
            time.Sleep(5 * time.Second) // 等待 5 秒后重试
        } else {
            return nil // 连接正常
        }
    }
    
    return errors.New("DB reconnect failed after max retry")
}

// 定时检查（每 30 秒）
func StartDBHealthChecker(svcCtx *svc.ServiceContext) {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            reconnector := &DBReconnector{db: svcCtx.DB, dsn: svcCtx.Config.DataSource, maxRetry: 3}
            reconnector.PingAndReconnect()
        }
    }()
}
```

---

#### 建议 4：异地容灾备份（可选）

**现状**：单机房部署

**优化建议**：异地主从同步

```
主库：北京机房（生产流量）
从库：上海机房（灾备）
    ↓ PostgreSQL Stream Replication
实时同步
```

**PostgreSQL 主从配置**：
```sql
-- 主库（北京）
ALTER SYSTEM SET wal_level = replica;
ALTER SYSTEM SET max_wal_senders = 3;

-- 从库（上海）
PRIMARY_CONNINFO = 'host=beijing-db port=5432 user=replicator'
```

---

#### 建议 5：服务注册与发现（Consul）

**现状**：
- tracking-api 直连 tracking-srv（硬编码 IP）
- 服务扩容困难

**优化建议**：启用 Consul 服务发现

```yaml
# tracking-srv/etc/tracking.yaml（新增 Consul 配置）
Consul:
  Host: 127.0.0.1:8500
  Key: tracking.rpc
  Meta:
    service: tracking-srv
    version: 1.0.0
```

**客户端配置**：
```yaml
# tracking-api/etc/tracking-api-api.yaml
TrackingRpc:
  Endpoints:
    - 127.0.0.1:8500  # Consul 地址
  Discovery: consul   # 启用服务发现
```

**优势**：
- 服务动态扩缩容
- 自动故障转移
- 健康检查集成

---

### 3.3 高可用性优化总表

| 建议 | 等级 | 实施步骤 | 预计工作量 |
|------|------|---------|-----------|
| 降级逻辑 | 高 | Redis 缓存 + Asynq 重试 | 2 小时 |
| 限流熔断 | 高 | Sentinel 集成 + 规则配置 | 2 小时 |
| DB 重连机制 | 中 | 定时 Ping + 自动重连 | 1 小时 |
| 异地容灾 | 低 | PostgreSQL 主从同步（可选） | 3 小时 |
| Consul 服务发现 | 中 | Consul 配置 + 动态发现 | 2 小时 |

---

## 四、综合评分与优化优先级

### 4.1 三维度评分

| 维度 | 当前评分 | 优化后预期评分 | 差距 |
|------|---------|---------------|------|
| **可维护性** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | +1 |
| **可观测性** | ⭐⭐ | ⭐⭐⭐⭐⭐ | +3 |
| **高可用性** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | +2 |

---

### 4.2 优化实施优先级（分 3 周）

#### 第 1 周：可维护性 + 高可用性（基础设施）

| Day | 任务 | 工作量 |
|-----|------|--------|
| **Day 1** | 新增 Service 层（聚合 DAO） | 2h |
| **Day 2** | 新增 DTO 层（Model → DTO） | 1h |
| **Day 3** | 新增 Constants 层（状态码） | 1h |
| **Day 4** | 降级逻辑实现（Redis + Asynq） | 2h |
| **Day 5** | 健康检查增强（依赖检查） | 1h |

---

#### 第 2 周：可观测性（监控体系）

| Day | 任务 | 工作量 |
|-----|------|--------|
| **Day 1** | 结构化日志（JSON + TraceID） | 1h |
| **Day 2** | Prometheus 指标暴露 | 2h |
| **Day 3** | Grafana Dashboard 配置 | 1h |
| **Day 4** | OpenTelemetry 集成（链路追踪） | 3h |
| **Day 5** | 告警规则配置（AlertManager） | 1h |

---

#### 第 3 周：高可用性进阶（容灾 + 限流）

| Day | 任务 | 工作量 |
|-----|------|--------|
| **Day 1** | Sentinel 限流集成 | 2h |
| **Day 2** | DB 重连机制实现 | 1h |
| **Day 3** | Consul 服务发现配置 | 2h |
| **Day 4** | 异地容灾方案设计（可选） | 3h |
| **Day 5** | 全链路压测验证 | 2h |

---

## 五、关键技术选型建议

### 5.1 可观测性技术栈

| 技术 | 用途 | 状态 |
|------|------|------|
| **Prometheus** | 指标收集 + 存储 | 推荐集成 |
| **Grafana** | 监控可视化 | 推荐集成 |
| **OpenTelemetry** | 链路追踪 | 推荐集成 |
| **Jaeger** | Trace UI 展示 | 推荐集成 |
| **AlertManager** | 告警管理 | 推荐集成 |

---

### 5.2 高可用性技术栈

| 技术 | 用途 | 状态 |
|------|------|------|
| **Sentinel** | 限流熔断 | 推荐集成 |
| **Consul** | 服务发现 | 推荐集成 |
| **Redis** | 降级缓存 | 已集成 |
| **Asynq** | 失败重试 | 已集成 |
| **PostgreSQL Stream Replication** | 异地容灾 | 可选 |

---

### 5.3 开发辅助技术栈

| 技术 | 用途 | 状态 |
|------|------|------|
| **Swagger/OpenAPI** | API 文档生成 | 推荐集成 |
| **Mockery** | Mock 测试生成 | 推荐集成 |
| **Air** | 热重载开发工具 | 推荐集成 |
| **GolangCI-Lint** | 代码静态检查 | 推荐集成 |

---

## 六、总结与下一步

### 6.1 当前架构健康度

```
┌──────────────────────────────────────┐
│         架构健康度评估                │
├──────────────────────────────────────┤
│                                      │
│  可维护性：⭐⭐⭐⭐ (80/100)          │
│  可观测性：⭐⭐     (40/100)          │
│  高可用性：⭐⭐⭐   (60/100)          │
│                                      │
│  总评分：    ⭐⭐⭐ (60/100)          │
│                                      │
│  优化后预期：⭐⭐⭐⭐⭐ (95/100)       │
│                                      │
└──────────────────────────────────────┘
```

---

### 6.2 核心问题优先级

| 问题 | 优先级 | 影响 |
|------|--------|------|
| **无监控指标** | P0（最高） | 无法发现性能问题、故障无感知 |
| **无降级逻辑** | P0（最高） | DB 故障导致数据丢失 |
| **无限流保护** | P1（高） | 云途推送过载导致服务崩溃 |
| **无链路追踪** | P1（高） | 故障定位困难 |
| **无 DTO 层** | P2（中） | Model 直接暴露，安全隐患 |

---

### 6.3 下一步行动建议

**优先实施（P0）**：
1. Prometheus 指标暴露（Day 2-3）
2. 降级逻辑实现（Day 4）
3. 健康检查增强（Day 5）

**次要实施（P1）**：
4. OpenTelemetry 集成（Day 4）
5. Sentinel 限流（Day 1）

**可选实施（P2）**：
6. Service/DTO 层重构
7. 异地容灾方案

---

**准备好开始优化时告诉我，我来帮你实施第一个优先级任务（Prometheus 指标暴露）。**