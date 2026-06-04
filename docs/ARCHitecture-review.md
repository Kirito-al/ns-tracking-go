# ns-tracking-go 架构评估报告

> 项目：Go-Zero YunExpress Tracking Webhook Service  
> 评估时间: 2026-06-03  
> 评估维度： 可维护性、 可观测性  高可用性

---

## 一、可维护性评估（代码结构清晰，业务扩展性)

### 1.1 当前架构分析

#### 目录结构
```
D:\gohome\demo1-gozero\ns-tracking-go
├── common/                    # 公共模块（跨服务共享)
│   ├── consts/                 # 常量定义
│   ├── errors/                 # 错误码定义
│   └── utils/                 # 工具函数
│
├── tracking-api/                 # HTTP API 服务
│   ├── internal/
│   │   ├── config/             # 配置管理
│   │   ├── handler/              # HTTP 处理器
│   │   │   ├── admin/          # 管理后台 API
│   │   │   ├── health/         # 健康检查
│   │   │   ├── public/         # 公开查询 API
│   │   │   ├── tracking/      # 单号查询
│   │   │   ├── webhook/      # Webhook 接收
│   │   ├── logic/               # 业务逻辑
│   │   │   ├── admin/
│   │   │   ├── public/
│   │   │   ├── tracking/
│   │   │   ├── webhook/
│   │   ├── middleware/           # 中间件
│   │   ├── svc/                 # 依赖注入
│   │   ├── types/                # 类型定义
│   │   ├── utils/                # 工具类（签名验证）
│   └ tracking-api.go              # 服务入口
│
├── tracking-srv/                 # gRPC 服务
│   ├── cmd/
│   │   ├── worker/             # Asynq Worker 进程
│   ├── internal/
│   │   ├── cache/               # 缓存层
│   │   ├── config/              # 配置管理
│   │   ├── dao/                 # 数据访问层
│   │   ├── event/               # 事件系统（新增）
│   │   │   ├── define/          # 事件定义
│   │   │   ├── dispatcher/      # 事件分发器
│   │   │   ├── listener/        # 事件监听器
│   │   ├── logic/               # 业务逻辑
│   │   ├── model/               # 数据模型
│   │   ├── queue/               # Asynq 队列（新增）
│   │   │   ├── config/
│   │   │   ├── handler/
│   │   │   ├── tasks/
│   │   ├── server/              # gRPC 服务实现
│   │   ├── svc/                 # 依赖注入
│   ├── tracking/                 # Proto 生成的 Go 代码
│   └ tracking.go              # gRPC 服务入口
│   └ tracking.proto           # Proto 定义
│
├── database.sql                # 数据库表结构
├── docker-compose.yml            # 中间件配置
```

#### 优点 ✅

| 项目 | 评价 | 说明 |
|------|------|------|
| **分层清晰** | ✅ 优秀 | Handler-Logic-DAO-Model 五层分离，符合 DDD 规范 |
| **职责单一** | ✅ 优秀 | 每个模块职责明确（handler 只负责 HTTP， logic 只负责业务）|
| **命名规范** | ✅ 优秀 | 文件名清晰（如 `webhook_logic.go` `upsert_logic.go`）|
| **依赖注入** | ✅ 优秀 | `ServiceContext` 统一管理依赖（DB、Redis、DAO、 Dispatcher） |
| **配置集中** | ✅ 优秀 | 配置在 `config.go` 统一管理（DatabaseConfig、TimeoutConfig） |
| **事件系统** | ✅ 优秀 | `event/define/dispatcher/listener` 结构清晰，解耦入库和事件 |
| **异步队列** | ✅ 优秀 | Asynq 瘦身独立 Worker 进程， |

#### 问题点 ⚠️

| 问题 | 等级 | 说明 |
|------|------|------|
| **缺少 Formatter 层** | 🔴 高 | 对标 Ruby `YunExpressTrackFormatter`，状态码映射、格式化逻辑缺失 |
| **缺少 Filter 层** | 🔴 高 | 对标 Ruby `TrackingEventFilter`，轨迹事件过滤逻辑缺失 |
| **缺少 Masking 层** | 🔴 高 | 对标 Ruby `TrackLocationMasking`，地点脱敏逻辑缺失 |
| **缺少 Service Resolver** | 🔴 高 | 对标 Ruby `TrackingService.resolve_service_class`，物流商路由逻辑缺失 |
| **缺少 Repository 层** | 🟡 中 | DAO 直接操作 DB，缺少 Repository 封装 |
| **缺少接口定义** | 🟡 中 | `types.go` 只有基础类型，缺少 DTO 定义 |
| **缺少错误处理** | 🟡 中 | 只有基础错误码，缺少统一错误体系 |

---

### 1.2 可维护性评分

| 维度 | 评分 | 说明 |
|------|------|------|
| **代码清晰度** | 8/10 | 分层清晰，命名规范 |
| **扩展性** | 6/10 | 缺少 Formatter/Filter/Masking 层 |
| **复刻 Ruby 模式** | 5/10 | 只完成核心入库逻辑 |
| **模块化** | 7/10 | 事件/队列模块化良好 |
| **依赖管理** | 8/10 | ServiceContext 注入优秀 |
| **配置管理** | 8/10 | DatabaseConfig/TimeoutConfig 配置化 |

---

### 1.3 可维护性优化建议

#### 🔴 高优先级（必须补充）

**1. 补充 Formatter 层（状态码映射）**

```go
// tracking-srv/internal/formatter/track_formatter.go
type TrackFormatter struct{}

// 状态码映射（完全对标 Ruby）
var StatusCodes = map[string]string{
    "0":   "NotFound",
    "10":  "InfoReceived",
    "20":  "InTransit",
    "30":  "AvailableForPickup",
    "40":  "DeliveryFailure",
    "50":  "Delivered",
    "60":  "Exception",
    "70":  "Expired",
    "80":  "Exception",
    "90":  "Exception_Returned",
    "100": "Exception_Cancel",
}

// PackageState 映射（完全对标 Ruby）
var StatusCodesToPackageStates = map[string]string{
    "0":   "0", // Undefined
    "10":  "4", // Received
    "20":  "2", // InTransit
    "30":  "2", // InTransit
    "40":  "6", // DeliveryFailed
    "50":  "3", // Delivered
    "60":  "6", // Exception
    "70":  "6", // Expired
    "80":  "6", // Exception
    "90":  "7", // Returned
    "100": "5", // Canceled
}

func (f *TrackFormatter) Format(payload YunExpressPayload) map[string]interface{} {
    // 1. 状态码映射
    statusCode := payload.TrackingStatus
    statusText := StatusCodes[statusCode]
    packageState := StatusCodesToPackageStates[statusCode]
    
    // 2. 格式化 detail
    detail := map[string]interface{}{
        "oriNumber":   payload.TrackingNumber,
        "oriChannel": payload.ProviderName,
        "status":      statusText,
        "events":      f.formatEvents(payload.OrderTrackingDetails),
    }
    
    return detail
}
```

---

**2. 补充 Filter 层（事件过滤）**

```go
// tracking-srv/internal/filter/event_filter.go
type EventFilter struct{}

// 对标 Ruby TrackingEventFilter
func (f *EventFilter) FilterEvents(events []TrackingEvent) []TrackingEvent {
    filtered := []TrackingEvent{}
    
    for _, event := range events {
        // 1. 过滤重复事件（相同时间 + 地点）
        if f.isDuplicate(event, filtered) {
            continue
        }
        
        // 2. 过滤无效事件（空内容）
        if event.ProcessContent == "" {
            continue
        }
        
        // 3. 过滤 Counterfeit 关键字
        if strings.Contains(event.ProcessContent, "Counterfeit") {
            continue
        }
        
        filtered = append(filtered, event)
    }
    
    return filtered
}
```

---

**3. 补充 Masking 层（地点脱敏）**

```go
// tracking-srv/internal/mask/location_mask.go
type LocationMasker struct{}

// 对标 Ruby TrackLocationMasking
func (m *LocationMasker) MaskLocation(resp map[string]interface{}, countryCode string) {
    details := resp["Item"]["OrderTrackingDetails"]
    
    for _, detail := range details {
        location := detail["ProcessLocation"]
        
        // 只保留目的国的 location（如 NL）
        if !strings.Contains(location, countryCode) {
            detail["ProcessLocation"] = ""  // 其他国家置空
        }
        
        // 移除中国省份信息
        detail["ProcessContent"] = m.removeChineseProvinces(detail["ProcessContent"])
    }
    
    return resp
}
```

---

**4. 补充 Service Resolver（物流商路由）**

```go
// tracking-srv/internal/service/service_resolver.go
type ServiceResolver struct{}

// 对标 Ruby TrackingService.resolve_service_class
func (r *ServiceResolver) Resolve(trackingNumber string) string {
    // 1. 云途单号判断
    if strings.HasPrefix(trackingNumber, "YT") {
        return "yunexpress"
    }
    
    // 2. WS 单号判断
    if strings.HasPrefix(trackingNumber, "WS") {
        return "yunexpress"
    }
    
    // 3. 其他物流商判断（17Track/SY/YW）
    // ...
    
    return "17track"  // 默认
}
```

---

#### 🟡 中优先级（建议补充）

**5. 添加 Repository 层（封装 DAO）**

```go
// tracking-srv/internal/repository/tracking_repository.go
type TrackingRepository interface {
    Upsert(ctx context.Context, trackingNumber string, detail json.RawMessage) error
    FindByTrackingNumber(ctx context.Context, trackingNumber string) (*TrackingDetail, error)
    FindLatest(ctx context.Context, trackingNumber string) (*TrackingDetail, error)
}

type TrackingRepositoryImpl struct {
    dao *dao.TrackingDAO
}

func (r *TrackingRepositoryImpl) Upsert(...) error {
    // 封装 DAO 的 Upsert 逻辑
    return r.dao.Upsert(...)
}
```

---

**6. 添加 DTO 定义层（统一数据传输对象）**

```go
// tracking-srv/internal/dto/tracking_dto.go
type UpsertRequest struct {
    TrackingNumber string `json:"tracking_number"`
    Detail          string `json:"detail"`
    Status          int32  `json:"status"`
    ServiceClass    string `json:"service_class"`
    SyncedAt        int64  `json:"synced_at"`
}

type UpsertResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}
```

---

**7. 补充错误处理体系**

```go
// common/errors/errors.go
type ErrorCode int

const (
    ErrNotFound         ErrorCode = 1001
    ErrInvalidSignature ErrorCode = 1002
    ErrDatabaseError    ErrorCode = 1003
    ErrInvalidParameter ErrorCode = 1004
)

type BusinessError struct {
    Code    ErrorCode
    Message string
}

func (e *BusinessError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}
```

---

### 1.4 扩展性设计

#### 新增物流商扩展流程

```
新增物流商（如 17Track）：
  1. 创建 formatter/seventeen_track_formatter.go
  2. 创建 service/seventeen_track_service.go
  3. 在 service_resolver.go 添加判断逻辑
  4. 在 config.go 添加 Token 配置
  5. 在 database.sql 添加表（如 tracking_17track_details）
```

**扩展成本**：低（只需新增 Formatter + Service， 无需修改核心架构）

---

## 二、可观测性评估（运维体系建设）

### 2.1 当前状态

#### 监控缺失 ⚠️
| 项目 | 状态 | 说明 |
|------|------|------|
| **Metrics 指标** | ❌ 无 | 缺少 Prometheus 指标暴露 |
| **Trace 链路追踪** | ❌ 无 | 缺少 OpenTelemetry 集成 |
| **Health Check** | ✅ 基础 | 有 `/health` 接口，但缺少详细状态 |
| **Log 集中化** | ❌ 无 | 终端日志，缺少文件日志 + Lumberjack |
| **Alert 告警** | ❌ 无 | 缺少 Grafana/PrometheusAlertManager 告警规则 |
| **Dashboard 监控面板** | ❌ 无 | 缺少 Grafana Dashboard |
| **Business Metrics** | ❌ 无 | 缺少业务指标（缓存命中率、Webhook 成功率） |

---

### 2.2 关键技术指标（缺失）

| 指标类别 | 指标名称 | 重要性 | 说明 |
|----------|----------|---------|------|
| **核心指标** | HTTP QPS | 🔴 高 | Webhook 接收频率 |
| **核心指标** | DB 连接数 | 🔴 高 | PostgreSQL 连接池使用率 |
| **核心指标** | Redis 连接数 | 🟡 中 | Redis 连接数/命中率 |
| **核心指标** | gRPC 调用延迟 | 🔴 高 | tracking-api → tracking-srv 延迟 |
| **核心指标** | DB 查询延迟 | 🔴 高 | Upsert 查询延迟（P99） |
| **核心指标** | Asynq 任务队列长度 | 🟡 中 | Redis 队列积压 |
| **核心指标** | Worker 处理速率 | 🟡 中 | Asynq Worker 处理速度 |
| **技术指标** | 错误率 | 🔴 高 | Webhook 签名验证失败率 |
| **技术指标** | 重试率 | 🟡 中 | Asynq 任务重试次数 |
| **技术指标** | GC 压力 | 🟢 低 | Go GC 压力（内存使用） |

---

### 2.3 关键业务指标（缺失）

| 指标名称 | 重要性 | 说明 |
|----------|---------|------|
| **Webhook 接收成功率** | 🔴 高 | 成功入库比例（目标 99%+） |
| **缓存命中率** | 🔴 高 | Ruby 读 tracking_details 成功率（目标 90%+） |
| **状态码分布** | 🟡 中 | 各状态码占比（Delivered/InTransit/Exception） |
| **单号渠道分布** | 🟡 中 | 云途/GE-云途 单号占比 |
| **同步延迟分布** | 🔴 高 | 云途推送 → 入库延迟（目标 <5 分钟） |
| **轨迹事件数量分布** | 🟢 低 | 平均每个单号的轨迹事件数 |
| **单号查询热度** | 🟡 中 | Top 100 单号查询频率 |

---

### 2.4 可观测性评分

| 维度 | 评分 | 说明 |
|------|------|------|
| **日志系统** | 3/10 | 只有终端日志，缺少文件日志和集中化 |
| **指标系统** | 0/10 | 完全缺失 Prometheus/OpenTelemetry |
| **链路追踪** | 0/10 | 完全缺失 OpenTelemetry Trace |
| **健康检查** | 5/10 | 有基础接口，缺少详细状态 |
| **告警系统** | 0/10 | 完全缺失告警规则 |
| **业务指标** | 0/10 | 完全缺失业务指标统计 |

---

### 2.5 可观测性优化建议

#### 🔴 高优先级（必须补充）

**1. 添加 Prometheus 指标暴露**

```go
// tracking-srv/internal/metrics/metrics.go
import "github.com/prometheus/client_golang/prometheus"

var (
    UpsertTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tracking_upsert_total",
            Help: "Total number of upsert operations",
        },
        []string{"status", "service_class"},
    )
    
    UpsertLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "tracking_upsert_latency_seconds",
            Help:    "Upsert operation latency",
            Buckets: []float64{.01, .05, .1, .5, 1, 5, 10},
        },
        []string{"service_class"},
    )
    
    WebhookReceived = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "webhook_received_total",
            Help: "Total number of webhook received",
        },
    )
    
    WebhookSignatureFailed = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "webhook_signature_failed_total",
            Help: "Total number of webhook signature validation failed",
        },
    )
)
```

---

**2. 添加 OpenTelemetry 链路追踪**

```go
// tracking-srv/internal/trace/trace.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func StartUpsertTrace(ctx context.Context, trackingNumber string) (context.Context, trace.Span) {
    tracer := otel.Tracer("tracking-srv")
    ctx, span := tracer.Start(ctx, "UpsertTracking",
        trace.WithAttributes(
            attribute.String("tracking_number", trackingNumber),
        ),
    )
    return ctx, span
}
```

---

**3. 添加 Zap + Lumberjack 文件日志**

```go
// tracking-srv/internal/log/logger.go
import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger() *zap.Logger {
    writer := &lumberjack.Logger{
        Filename:   "logs/tracking-srv.log",
        MaxSize:    100, // MB
        MaxBackups: 3,
        MaxAge:     7,   // days
        Compress:   true,
    }
    
    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(writer),
        zap.InfoLevel,
    )
    
    return zap.New(core)
}
```

---

**4. 添加 Grafana Dashboard**

```yaml
# monitoring/grafana/dashboards/tracking-dashboard.json
{
  "title": "Tracking Service Dashboard",
  "panels": [
    {
      "title": "Upsert QPS",
      "type": "graph",
      "targets": [
        {
          "expr": "rate(tracking_upsert_total[5m])"
        }
      ]
    },
    {
      "title": "Upsert Latency (P99)",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.99, tracking_upsert_latency_seconds_bucket)"
        }
      ]
    },
    {
      "title": "Webhook Success Rate",
      "type": "gauge",
      "targets": [
        {
          "expr": "webhook_received_total / (webhook_received_total + webhook_signature_failed_total)"
        }
      ]
    }
  ]
}
```

---

#### 🟡 中优先级（建议补充）

**5. 添加健康检查详细状态**

```go
// tracking-api/internal/handler/health/health_handler.go
type HealthStatus struct {
    Service   string `json:"service"`
    Status    string `json:"status"`
    DB        string `json:"db"`        // "connected" / "disconnected"
    Redis     string `json:"redis"`     // "connected" / "disconnected"
    Asynq     string `json:"asynq"`     // "running" / "stopped"
    Uptime    int64  `json:"uptime"`    // 服务启动时间（秒）
    Version   string `json:"version"`   // 服务版本
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
    status := HealthStatus{
        Service: "tracking-api",
        Status:  "healthy",
        DB:      h.checkDB(),
        Redis:   h.checkRedis(),
        Uptime:  time.Since(h.startTime).Seconds(),
        Version: "1.0.0",
    }
    
    json.NewEncoder(w).Encode(status)
}
```

---

**6. 添加业务指标统计**

```go
// tracking-srv/internal/metrics/business_metrics.go
func RecordBusinessMetrics(detail *model.TrackingDetail) {
    // 1. 状态码分布
    UpsertTotal.WithLabelValues(detail.StatusTxt, detail.ServiceClass).Inc()
    
    // 2. 缓存命中率（Redis）
    if detail.SyncedAt > time.Now().Unix()-300 {
        CacheHit.Inc()
    } else {
        CacheMiss.Inc()
    }
    
    // 3. 轨迹事件数量
    EventCount.Observe(float64(len(detail.Detail.OrderTrackingDetails)))
}
```

---

## 三、高可用性评估（服务稳定性、响应时效性）

### 3.1 当前状态

#### 服务稳定性 ⚠️
| 项目 | 状态 | 说明 |
|------|------|------|
| **多实例部署** | ❌ 无 | 单实例，无副本（Replicas） |
| **服务发现** | ❌ 无 | 直接连接（无 Consul/Etcd） |
| **负载均衡** | ❌ 无 | 无 Nginx/Kong 负载均衡 |
| **健康检查** | ✅ 基础 | 有 `/health` 但缺少 K8s 探针 |
| **故障转移** | ❌ 无 | 单实例故障无法自动切换 |
| **降级策略** | 🟡 部分 | Ruby Pull 降级（环境变量开关） |
| **熔断机制** | ❌ 无 | 缺少 Circuit Breaker |
| **限流机制** | ❌ 无 | 缺少 Rate Limiter |
| **重试机制** | ✅ Asynq | Asynq 内置重试（MaxRetry 3） |

---

#### 响应时效性 ⚠️
| 项目 | 当前延迟 | 目标延迟 | 说明 |
|------|----------|----------|------|
| **Webhook 接收** | <100ms | <50ms | HTTP 响应时间 |
| **gRPC Upsert** | 1~5s | <500ms | DB Upsert 延迟（优化 JSONB 查询） |
| **DB Upsert** | 1~5s | <200ms | PostgreSQL Upsert 延迟 |
| **Ruby 读缓存** | N/A | <50ms | Ruby 读 tracking_details |
| **整体链路** | 1~5s | <500ms | Webhook → DB 入库 |

---

### 3.2 高可用性评分

| 维度 | 评分 | 说明 |
|------|------|------|
| **多实例部署** | 0/10 | 单实例，无副本 |
| **服务发现** | 0/10 | 直接连接，无注册中心 |
| **负载均衡** | 0/10 | 无负载均衡 |
| **健康检查** | 5/10 | 有基础接口，缺少探针 |
| **故障转移** | 0/10 | 单实例无自动切换 |
| **降级策略** | 6/10 | 有 Pull 降级开关 |
| **熔断机制** | 0/10 | 完全缺失 |
| **限流机制** | 0/10 | 完全缺失 |
| **重试机制** | 8/10 | Asynq 内置重试良好 |
| **响应时效** | 6/10 | 延迟可接受，但可优化 |

---

### 3.3 高可用性优化建议

#### 🔴 高优先级（必须补充）

**1. 添加服务发现（Consul）**

```yaml
# tracking-srv/etc/tracking.yaml
Consul:
  Host: 127.0.0.1:8500
  Key: tracking.rpc
  Meta:
    service: tracking-srv
    version: 1.0.0
```

```go
// tracking-srv/internal/svc/servicecontext.go
import "github.com/zeromicro/go-zero/core/discov"

func NewServiceContext(c config.Config) *ServiceContext {
    // 注册到 Consul
    if c.Consul.Host != "" {
        discov.Register(c.Consul)
    }
}
```

---

**2. 添加多实例部署**

```yaml
# docker-compose.yml
tracking-srv:
  deploy:
    replicas: 3  # 3 个副本
    update_config:
      parallelism: 1
      delay: 10s
    restart_policy:
      condition: on-failure
      delay: 5s
      max_attempts: 3
```

---

**3. 添加负载均衡**

```yaml
# docker-compose.yml
nginx:
  image: nginx:alpine
  ports:
    - "8080:80"
  volumes:
    - ./nginx.conf:/etc/nginx/nginx.conf
  depends_on:
    - tracking-api

# nginx.conf
upstream tracking_api {
    server tracking-api-1:8081;
    server tracking-api-2:8081;
    server tracking-api-3:8081;
}

server {
    listen 80;
    location /webhook/ {
        proxy_pass http://tracking_api;
    }
}
```

---

**4. 添加熔断机制**

```go
// tracking-api/internal/middleware/circuit_breaker.go
import "github.com/afex/hystrix-go/hystrix"

func InitCircuitBreaker() {
    hystrix.ConfigureCommand("webhook_upsert", hystrix.CommandConfig{
        Timeout:                5000,  // 5 秒超时
        MaxConcurrentRequests:  100,   // 最大并发
        ErrorPercentThreshold:  50,    // 50% 错误率触发熔断
    })
}

func (m *CircuitBreakerMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        hystrix.Do("webhook_upsert", func() error {
            next.ServeHTTP(w, r)
            return nil
        }, func(err error) error {
            w.WriteHeader(http.StatusServiceUnavailable)
            return nil
        })
    })
}
```

---

**5. 添加限流机制**

```go
// tracking-api/internal/middleware/rate_limiter.go
import "golang.org/x/time/rate"

func (m *RateLimiterMiddleware) Handle(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 100)  // 100 QPS
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

#### 🟡 中优先级（建议补充）

**6. 添加 DB Upsert 性能优化**

```go
// tracking-srv/internal/dao/tracking_dao.go
// 使用 Batch Upsert（批量插入）
func (d *TrackingDAO) BatchUpsert(ctx context.Context, details []*model.TrackingDetail) error {
    // 单条 SQL 批量插入（比循环调用快 10x）
    err := d.db.WithContext(ctx).Exec(`
        INSERT INTO tracking_details (tracking_number, detail, status, ...)
        VALUES ($1, $2, $3, ...), ($1, $2, $3, ...), ...
        ON CONFLICT (tracking_number) DO UPDATE SET ...
    `, ...).Error
    
    return err
}
```

---

**7. 添加 JSONB 查询优化（索引）**

```sql
-- database.sql
-- 添加 GIN 索引（加速 JSONB 查询）
CREATE INDEX idx_tracking_details_detail_gin ON tracking_details USING GIN (detail);

-- 添加 status 狀態索引
CREATE INDEX idx_tracking_details_status ON tracking_details(status);

-- 添加 synced_at 时间索引
CREATE INDEX idx_tracking_details_synced_at ON tracking_details(synced_at);
```

---

## 四、优先级排序（行动计划）

### 🔴 必须补充（Phase 2）
1. Formatter 层（状态码映射）
2. Filter 层（事件过滤）
3. Prometheus 指标暴露
4. Zap + Lumberjack 文件日志
5. 服务发现（Consul）
6. 多实例部署（3 副本）
7. DB Upsert 性能优化（索引）

### 🟡 建议补充（Phase 3）
8. OpenTelemetry 链路追踪
9. Grafana Dashboard
10. 熔断机制（Circuit Breaker）
11. 限流机制（Rate Limiter）
12. Masking 层（地点脱敏）
13. Service Resolver（物流商路由）

### 🟢 可选补充（Phase 4）
14. Repository 层（封装 DAO）
15. DTO 定义层
16. 错误处理体系
17. 业务指标统计
18. 健康检查详细状态

---

## 五、总结

### 当前架构评分

| 维度 | 总分 | 说明 |
|------|------|------|
| **可维护性** | 6.5/10 | 分层清晰，但缺少 Formatter/Filter/Masking 层 |
| **可观测性** | 1.5/10 | 只有基础健康检查，完全缺失监控体系 |
| **高可用性** | 2.5/10 | 单实例，无服务发现，无负载均衡 |

### 优化后预期评分

| 维度 | 优化后总分 | 说明 |
|------|-----------|------|
| **可维护性** | 9.0/10 | 补充 Formatter/Filter/Masking 层，完全对标 Ruby |
| **可观测性** | 8.0/10 | 补充 Prometheus/OpenTelemetry/Grafana |
| **高可用性** | 8.5/10 | 补充服务发现/多实例/负载均衡/熔断 |

---

**建议立即启动 Phase 2，从 Formatter 层开始补充。**