# YunExpress Webhook Service

> 跨境云途物流 Webhook 接收服务，基于 go-zero 微服务框架实现。
  项目目标：对接云途物流 Webhook 推送，自动接收运单轨迹、验签、格式化、幂等入库，对外提供轨迹查询 HTTP 接口；
          替代原有 Ruby 定时轮询方案，减少接口调用损耗、实时获取轨迹。
> 已落地：云途 Webhook 接收、HMAC-SHA256 验签、数据格式化、gRPC 微服务落库、运单查询接口、PG 幂等 UPSERT
> 迭代规划：事件驱动（美国收货自动发邮件）
> 技术栈：Go-Zero、GORM、PostgreSQL、Redis、Consul、Zap 日志、Lumberjack viper日志分割
---

## 一、项目架构说明

### 1.1 微服务分层架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     云途物流 Webhook 推送                        │
│                   POST /webhook/yunexpress/tracking              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                    tracking-api (HTTP API)                       │
│                      监听端口：0.0.0.0:8082                        │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ Handler 层：验签 + 解析 JSON + 路由分发                        ││
│  └─────────────────┬───────────────────────────────────────────┘│
│                    ↓ Logic 层：业务逻辑处理                        │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ formatPayload() + gRPC Client 调用                            ││
│  └─────────────────┬───────────────────────────────────────────┘│
└────────────────────┼─────────────────────────────────────────────┘
                     │ gRPC (127.0.0.1:50051)
                     ↓
┌─────────────────────────────────────────────────────────────────┐
│                    tracking-srv (gRPC Service)                   │
│                      监听端口：127.0.0.1:50051                     │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ Logic 层：Upsert + GetTracking + MarkError                    ││
│  └─────────────────┬───────────────────────────────────────────┘│
│                    ↓ DAO 层：GORM 数据访问                         │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ INSERT ... ON CONFLICT DO UPDATE (PostgreSQL Upsert)         ││
│  └─────────────────┬───────────────────────────────────────────┘│
└────────────────────┼─────────────────────────────────────────────┘
                     │ PostgreSQL (127.0.0.1:5432)
                     ↓
┌─────────────────────────────────────────────────────────────────┐
│                    PostgreSQL Database                           │
│                  数据库：ns_admin_webhook_development            │
│  表结构：tracking_details / tracking_logs / tracking_cache_logs │
└─────────────────────────────────────────────────────────────────┘
```

---

### 1.2 技术栈清单

| 技术栈 | 版本 | 用途 |
|--------|------|------|
| **go-zero** | v1.10.2 | 微服务框架（REST + gRPC） |
| **GORM** | v1.31.1 | ORM 数据访问层 |
| **PostgreSQL** | v14 | 关系型数据库（JSONB 存储） |
| **Redis** | v6 | 缓存 + 分布式锁 |
| **Consul** | v1.14 | 服务发现 + 配置中心（可选） |
| **Asynq** | - | 异步任务队列（规划中） |
| **zap** | v1.24.0 | 结构化日志 |
| **lumberjack** | v2.2.1 | 日志轮转 |
| **viper** | v1.21.0 | 配置管理 |

---

### 1.3 数据流转链路

```
云途 Webhook 推送             (云途在运单轨迹发生更新后，主动 HTTP POST,
                             请求 Header 携带HMAC-SHA256 签名，Body 是云途原生 JSON 报文)
    ↓
Handler 层：验签（HMAC-SHA256）职责只做HTTP 收发包、安全校验，无业务、不操作数据库：
                             读取请求 Body 完整报文 + Header 内加密签名串；
                             使用项目约定密钥做 HMAC-SHA256 验签：
                             验签失败：直接返回错误码，拦截非法请求（防恶意伪造、报文篡改）；
                             验签通过：把原始 Body 向下传给 Logic。
                             规范：Handler 禁止复杂业务，保证接口快速响应。
    ↓
Logic 层：formatPayload() 格式化数据 报文结构转换，解决云途原始 JSON 结构和内部实体不一致问题（原生外层无 Tracking/Item，统一规整为项目内部结构体）；
                         JSON 序列化后，通过gRPC 客户端远程调用 tracking-srv 的UpsertTracking方法，入参：运单号、轨迹详情 detail、单据状态 status；
                         API 层不直连 PG，数据库操作全部下沉 RPC 服务，实现内外网隔离。
    ↓
gRPC 调用：UpsertTracking(trackingNumber, detail, status) tracking-srv (gRPC 服务) 接收远程入参，进入自身内部逻辑，调用 DAO 层入库。
    ↓
DAO 层：GORM                使用 PG 专属语法：INSERT ... ON CONFLICT DO UPDATE
                           数据库无此 trackingNumber（唯一索引）→ 新增入库；
                           已存在同运单号 → 触发冲突，覆盖更新最新轨迹；
                           单 SQL 原子执行，规避并发重复推送导致重复创建数据。
                           入库成功时同步赋值：syncedAt=当前UTC时间戳、errorMessage=""；入库异常则 syncedAt=0、填入错误信息。
    ↓
PostgreSQL：tracking_details 表入库 (字段：运单号、detail (jsonb 原始轨迹)、status、syncedAt 同步时间、errorMessage 异常信息。)
    ↓
返回 HTTP 200 OK（云途判定推送成功） RPC 执行完毕逐层结果回传 → Logic → Handler；
                                Handler 返回HTTP 200；
                                云途收到 200 代表推送成功，不会重复重试推送；如果返回非 200，云途会按配置策略多次补发。
```

---

## 二、目录说明

### 2.1 项目根目录结构

```
demo1-gozero/
├── tracking.api               # HTTP API 定义文件（goctl 生成标准）
├── database.sql               # PostgreSQL 建表 SQL + 测试数据
├── docker-compose.yml         # Docker Compose 配置（PostgreSQL + Redis + Consul）
├── README.md                  # 项目文档
│
├── common/                    # 公共模块（跨服务共享）
├── pkg/                       # 第三方接口封装（规划中）
├── tracking-api/              # HTTP API 服务（go-zero rest 框架）
└── tracking-srv/              # gRPC 服务（go-zero rpc 框架）
```

---

### 2.2 common 公共模块

```
common/
├── consts/                    # 常量定义
│   ├── status_codes.go       # 状态码常量（0-1001）
│   └── channel_prefix.go     # 渠道前缀（GE-云途、GEYT 等）
│
├── errors/                    # 错误码定义
│   ├── code.go               # 错误码枚举
│   └── message.go            # 错误消息映射
│
└── utils/                     # 公共工具函数
    ├── signer.go             # HMAC-SHA256 签名验证
    ├── time.go               # 时间戳转换（Unix ↔ ISO8601）
    └── formatter.go          # JSON 格式化工具
```

**作用说明**：

| 目录 | 用途 | 使用场景 |
|------|------|---------|
| `consts` | 定义全局常量 | 状态码映射、渠道判断 |
| `errors` | 统一错误码 | Handler 层返回标准错误 |
| `utils` | 公共工具 | 签名验证、时间处理 |

---

### 2.3 pkg 第三方接口封装（规划中）

```
pkg/
├── yunexpress/                # 云途物流 API 封装
│   ├── client.go             # HTTP Client 初始化
│   ├── tracking.go           # GetTrackInfo 接口封装
│   └── auth.go               # Token 认证管理
│
├── usps/                     # USPS（美国邮政）接口封装
│   ├── client.go             # USPS API Client
│   └── tracking.go           # USPS 轨迹查询
│
└── pay/                    # 第三方支付渠道统一SDK封装
│   ├── paypal              # PayPal 跨境支付接口封装
│   └── stripe              # Stripe 欧美信用卡支付封装
│
└── dhl/                       # DHL 接口封装（扩展预留）
    ├── client.go             # DHL API Client
    └── tracking.go           # DHL 轨迹查询
    
```

**作用说明**：

| 目录 | 用途 | 使用场景 |
|------|------|---------|
| `yunexpress` | 云途 API 封装 | 轮询模式拉取轨迹（旧方案） |
| `usps` | USPS API 封装 | 尾程轨迹查询（美国境内） |
| `dhl` | DHL API 封装 | 未来扩展其他物流商 |
| `pay` | 跨境支付SDK统一封装 | 外贸订单运费收款、线上支付结算 |
| ├ `paypal` | PayPal支付接口封装 | 欧美买家常用在线付款 |
| └ `stripe` | Stripe信用卡支付封装 | 海外信用卡一键扣款 |

**当前状态**：**未实现**（Webhook 推送模式已替代轮询）。

---

### 2.4 event 事件包（规划中 - 企业级标准结构）

> Event 事件包标准结构（解耦、可扩展、可维护，go-zero 通用）
> 用途：事件定义、事件分发、事件监听、异步处理（如：美国发货发邮件）

**目录结构**：

```
internal/
├── event/                     # 事件根包
│   ├── define/               # 事件定义（所有事件结构体）
│   │   └── types.go         # 定义事件结构体：TrackingCreatedEvent
│   │
│   ├── dispatcher/           # 事件分发器（触发事件用）
│   │   └── dispatcher.go    # 事件分发核心（Publish/Subscribe）
│   │
│   ├── listener/             # 事件监听器（真正做事：发邮件、发短信、写日志）
│   │   ├── listener.go      # 监听器接口
│   │   ├── shipping_listener.go  # 美国发货监听（发邮件）
│   │   └── sms_listener.go  # 短信通知监听（预留）
│   │
│   └── register.go          # 统一注册监听器（启动时绑定）
```

**事件流程示例**（美国发货发邮件）：

```
1. 业务代码触发事件
   event.dispatcher.Dispatch(&TrackingCreatedEvent{TrackingNumber: "YT..."})

2. 事件分发器广播
   dispatcher 查找所有注册的监听器 → 调用 listener.Handle()

3. 监听器处理业务
   ShippingListener → 调用 SMTP 服务 → 发送邮件通知
```

**核心代码示例**：

```go
// internal/event/define/types.go
package define

type TrackingCreatedEvent struct {
    TrackingNumber string
    CountryCode    string
    Email          string
}

// internal/event/dispatcher/dispatcher.go
package dispatcher

type EventDispatcher struct {
    listeners map[string][]EventListener
}

func (d *EventDispatcher) Dispatch(event Event) {
    listeners := d.listeners[event.Type()]
    for _, listener := range listeners {
        go listener.Handle(event) // 异步处理
    }
}

// internal/event/listener/shipping_listener.go
package listener

type ShippingListener struct {
    mailService *mail.Service
}

func (l *ShippingListener) Handle(event Event) {
    e := event.(*define.TrackingCreatedEvent)
    if e.CountryCode == "US" {
        l.mailService.SendUSShippingNotification(e.Email)
    }
}

// internal/event/register.go
package event

func RegisterListeners() {
    dispatcher.Register("tracking.created", &ShippingListener{})
    dispatcher.Register("tracking.created", &LogListener{})
}
```

**作用说明**：

| 目录 | 用途 | 使用场景 |
|------|------|---------|
| `define` | 定义事件结构体 | 所有事件类型统一放置 |
| `dispatcher` | 事件分发器 | `Dispatch()` 触发事件 |
| `listener` | 事件监听器 | 真正执行业务逻辑（邮件、短信等） |
| `register.go` | 统一注册 | 启动时绑定监听器到事件类型 |

**当前状态**：**规划中**（当前 Webhook 模式已满足需求，未来扩展异步通知时使用）。

---

### 2.5 tracking-api HTTP API 服务

```
tracking-api/
├── tracking-api.go            # 服务入口（main 函数）
├── go.mod                     # Go 模块依赖
├── go.sum                     # 依赖版本锁定
│
├── bootstrap/                 # 启动引导（初始化流程）
│   └── app.go                # viper 配置初始化 + zap 日志初始化
│
├── etc/                       # 配置文件目录
│   └── tracking-api-api.yaml  # REST API 配置（端口、Redis、gRPC Client）
│
├── internal/                  # 内部实现（不对外暴露）
│   │
│   ├── config/                # 配置结构体定义
│   │   └── config.go         # Config struct（映射 YAML 配置）
│   │
│   ├── handler/               # HTTP 处理器（goctl 生成）
│   │   ├── routes.go         # 路由注册（/webhook, /internal/tracking, /health）
│   │   ├── webhook/          # Webhook Handler（验签 + 调用 Logic）
│   │   │   └── webhook_handler.go
│   │   ├── tracking/         # 查询 Handler
│   │   │   └── get_tracking_handler.go
│   │   ├── admin/            # 管理后台 Handler
│   │   │   ├── batch_query_handler.go
│   │   │   └── get_tracking_handler.go
│   │   ├── public/           # 公开 API Handler
│   │   │   └── query_handler.go
│   │   └── health/           # 健康检查 Handler
│   │       └── health_handler.go
│   │
│   ├── logic/                 # 业务逻辑层
│   │   ├── webhook/          # Webhook Logic（formatPayload + gRPC 调用）
│   │   │   └── webhook_logic.go
│   │   ├── tracking/         # 查询 Logic
│   │   │   └── get_tracking_logic.go
│   │   ├── admin/            # 管理后台 Logic
│   │   │   └── batch_query_logic.go
│   │   ├── public/           # 公开 API Logic
│   │   │   └── query_logic.go
│   │   └── health/           # 健康检查 Logic
│   │       └── health_logic.go
│   │
│   ├── middleware/            # 中间件（HTTP 拦截器）
│   │   ├── cors.go           # CORS 跨域处理
│   │   ├── auth.go           # JWT 认证（预留）
│   │   └── ratelimit.go      # 限流中间件
│   │
│   ├── svc/                   # 依赖注入容器（ServiceContext）
│   │   └── servicecontext.go  # gRPC Client + Redis + Config 初始化
│   │
│   ├── types/                 # HTTP 请求/响应结构体
│   │   └── types.go          # WebhookRequest, TrackingDetailDTO 等定义
│   │
│   └── utils/                 # 工具类（服务内部）
│       ├── hmac.go           # HMAC-SHA256 签名工具
│       └── formatter.go      # JSON 格式化工具
│
└── Dockerfile                 # Docker 构建文件（生产部署）
```

**核心目录说明**：

| 目录 | 用途 | 关键文件 |
|------|------|---------|
| `bootstrap` | 启动初始化 | `app.go`（viper + zap） |
| `etc` | 配置文件 | `tracking-api-api.yaml` |
| `handler` | HTTP 处理器 | `routes.go`（路由注册） |
| `logic` | 业务逻辑 | `webhook_logic.go`（核心逻辑） |
| `svc` | 依赖注入 | `servicecontext.go`（gRPC Client） |
| `types` | 数据结构 | `types.go`（请求/响应定义） |

---

### 2.6 tracking-srv gRPC 服务

```
tracking-srv/
├── tracking.go                # 服务入口（main 函数）
├── tracking.proto             # gRPC Proto 定义（生成 Go 代码）
├── go.mod                     # Go 模块依赖
├── go.sum                     # 依赖版本锁定
│
├── bootstrap/                 # 启动引导
│   └── app.go                # viper 配置初始化 + zap 日志初始化
│
├── etc/                       # 配置文件目录
│   └── tracking.yaml         # RPC 服务配置（端口、数据库、Redis）
│
├── tracking/                  # Proto 生成的 Go 代码（goctl 生成）
│   ├── tracking.pb.go        # Proto 消息定义
│   └── tracking_grpc.pb.go   # gRPC Client/Server 代码
│
├── internal/                  # 内部实现
│   │
│   ├── config/                # 配置结构体
│   │   └── config.go         # Config struct（数据库、Redis、Consul）
│   │
│   ├── svc/                   # 依赖注入容器
│   │   └── servicecontext.go  # GORM DB + Redis + DAO 初始化
│   │
│   ├── model/                 # 数据模型定义（GORM Model）
│   │   └── models.go         # TrackingDetail, TrackingLog 等结构体
│   │
│   ├── dao/                   # 数据访问层（GORM 实现）
│   │   └── tracking_dao.go   # Upsert、GetLatest、MarkError
│   │
│   ├── cache/                 # 缓存层（Redis）
│   │   └── tracking_cache.go # 缓存管理（4 小时 TTL）
│   │
│   ├── logic/                 # 业务逻辑层
│   │   ├── upsert_logic.go   # Upsert 处理逻辑
│   │   ├── get_tracking_logic.go  # 查询处理逻辑
│   │   └── mark_error_logic.go    # 错误标记逻辑
│   │
│   └── server/                # gRPC 服务实现
│       └── tracking_server.go # RegisterRpcServer 实现
│
└── Dockerfile                 # Docker 构建文件（生产部署）
```

**核心目录说明**：

| 目录 | 用途 | 关键文件 |
|------|------|---------|
| `bootstrap` | 启动初始化 | `app.go`（viper + zap） |
| `etc` | 配置文件 | `tracking.yaml` |
| `tracking` | Proto 生成代码 | `tracking.pb.go` + `tracking_grpc.pb.go` |
| `model` | 数据模型 | `models.go`（GORM Model） |
| `dao` | 数据访问层 | `tracking_dao.go`（Upsert 实现） |
| `cache` | 缓存层 | `tracking_cache.go`（Redis 缓存） |
| `logic` | 业务逻辑 | `upsert_logic.go`（核心逻辑） |
| `server` | gRPC 服务 | `tracking_server.go`（注册服务） |

---

## 三、环境部署指南

### 3.1 前置条件

| 工具 | 版本要求 | 说明 |
|------|---------|------|
| **Go** | 1.25+ | Go 语言运行环境 |
| **Docker** | 最新版 | Docker Desktop（Windows/Mac） |
| **PostgreSQL** | 14+ | 或通过 Docker 启动 |
| **Redis** | 6+ | 或通过 Docker 启动 |

---

### 3.2 启动中间件（Docker Compose）

```powershell
# 进入项目目录
cd D:\gohome\demo1-gozero

# 启动 PostgreSQL + Redis
docker-compose up -d postgres redis

# 查看服务状态
docker-compose ps

# 预期输出：
# NAME                STATUS    PORTS
# tracking-postgres   running   0.0.0.0:5432->5432/tcp
# tracking-redis      running   0.0.0.0:6379->6379/tcp
```

**可选：启用 Consul（服务发现）**

```powershell
# 取消 docker-compose.yml 第 15-29 行注释后启动
docker-compose up -d consul

# 访问 Consul UI：http://localhost:8500
```

---

### 3.3 初始化数据库

```powershell
# 方法 1：使用 psql 命令（需安装 PostgreSQL 客户端）
psql -h localhost -U postgres -d ns_admin_webhook_development -f database.sql

# 方法 2：使用 Docker 执行（推荐）
docker exec -i tracking-postgres psql -U postgres -d ns_admin_webhook_development < database.sql
```

**验证数据初始化**：

```powershell
docker exec -it tracking-postgres psql -U postgres -d ns_admin_webhook_development -c "SELECT tracking_number, status FROM tracking_details LIMIT 2;"
```

---

### 3.4 配置文件说明

#### tracking-api 配置（`tracking-api/etc/tracking-api-api.yaml`）

```yaml
Name: tracking-api
Host: 0.0.0.0
Port: 8082  # HTTP API 端口

# gRPC Client（直连模式）
TrackingRpc:
  Target: 127.0.0.1:50051  # tracking-srv 地址
  Timeout: 5000

# Redis 配置
RedisConf:
  Host: localhost:6379
  Pass: ""
  Db: 0

# 安全配置
YunExpressWebhookSecret: test-secret      # 云途 Webhook 签名 Secret
GEYunExpressWebhookSecret: test-ge-secret # GE 账号 Secret
YunExpressSkipSignature: true             # 开发环境跳过验签

# CORS 配置
CORS:
  AllowedOrigins:
    - "*"
  AllowedMethods:
    - GET
    - POST
```

#### tracking-srv 配置（`tracking-srv/etc/tracking.yaml`）

```yaml
Name: tracking.rpc
ListenOn: 127.0.0.1:50051  # gRPC 监听端口

# PostgreSQL 配置
DataSource: postgres://postgres:123456@127.0.0.1:5432/ns_admin_webhook_development?sslmode=disable

# Redis 配置
RedisConf:
  Host: localhost:6379
  Pass: ""
  Db: 0

# 服务发现（可选）
# Consul:
#   Host: 127.0.0.1:8500
#   Key: tracking.rpc
```

---

### 3.5 启动 Go 服务

**启动顺序（重要）**：

```
PostgreSQL/Redis（中间件） → tracking-srv（gRPC） → tracking-api（HTTP）
```

#### 启动 tracking-srv（gRPC 服务）

```powershell
cd D:\gohome\demo1-gozero\tracking-srv

# 安装依赖
go mod tidy

# 启动服务
go run tracking.go -f etc/tracking.yaml
```

**预期输出**：
```
{"level":"info","content":"PostgreSQL connected (GORM): postgres://..."}
{"level":"info","content":"Redis connected: localhost:6379"}
Starting rpc server at 127.0.0.1:50051...
```

---

#### 启动 tracking-api（HTTP API）

```powershell
cd D:\gohome\demo1-gozero\tracking-api

# 安装依赖
go mod tidy

# 启动服务
go run tracking-api.go -f etc/tracking-api-api.yaml
```

**预期输出**：
```
Starting server at 0.0.0.0:8082...
```

---

#### 验证服务健康

```powershell
# 测试健康检查接口
curl http://localhost:8082/health

# 预期响应：
# {"code":200,"message":"healthy","data":{"service":"yunexpress-webhook","status":"ok"}}
```

---

## 四、接口调用文档

### 4.1 Webhook 接收接口

**接口路径**：`POST /webhook/yunexpress/tracking/normal`

**功能**：接收云途物流 Webhook 推送，验签后写入数据库。

---

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `trackingNumber` | string | ✅ | 云途运单号（主单号） |
| `wayBillNumber` | string | ✅ | 运单号（通常与 trackingNumber 相同） |
| `trackingStatus` | string | ✅ | 当前最新状态码 |
| `packageState` | string | ✅ | 包裹状态（2=运输中） |
| `orderTrackingDetails` | array | ✅ | 轨迹明细数组 |
| `providerName` | string | ❌ | 物流服务商名称 |
| `countryCode` | string | ❌ | 目的国代码（如：US） |
| `trackingNumber2` | string | ❌ | 尾程单号 |
| `lastMileCarrierName` | string | ❌ | 尾程承运商 |

---

#### orderTrackingDetails 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `processDate` | string | 扫描时间（ISO8601） |
| `processLocation` | string | 扫描地点 |
| `processContent` | string | 事件描述 |
| `trackingStatus` | string | 该节点状态码 |

---

#### 签名验证规则

**Header**：`X-YunExpress-Signature`

**算法**：HMAC-SHA256

**验证流程**：
```
1. 读取原始 Body（byte[]）
2. 从 Header 提取签名
3. 使用 Secret 计算 HMAC-SHA256
4. 比对签名是否匹配
5. 不匹配 → 返回 401 Unauthorized
```

**开发环境**：`YunExpressSkipSignature: true`（跳过验签）

**生产环境**：必须开启验签，Secret 从云途开发者平台获取。

---

#### 请求示例

```json
{
  "trackingNumber": "YT2606500704802225",
  "wayBillNumber": "YT2606500704802225",
  "trackingStatus": "20",
  "packageState": "2",
  "providerName": "云途物流",
  "countryCode": "US",
  "orderTrackingDetails": [
    {
      "processDate": "2025-05-20T08:00:00Z",
      "processLocation": "深圳仓库",
      "processContent": "快件电子信息已收到",
      "trackingStatus": "10"
    },
    {
      "processDate": "2025-05-21T10:30:00Z",
      "processLocation": "深圳口岸",
      "processContent": "快件已发出",
      "trackingStatus": "20"
    }
  ]
}
```

---

#### 响应示例（成功）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "success": true,
    "trackingNumber": "YT2606500704802225",
    "status": "20",
    "packageState": "2"
  }
}
```

---

#### 响应示例（验签失败）

```json
{
  "code": 401,
  "message": "Invalid HMAC signature",
  "data": "签名验证失败（HMAC-SHA256 不匹配）"
}
```

---

#### 状态码说明

| 状态码 | 说明 | PackageState |
|--------|------|--------------|
| `0` | NotFound | 0 |
| `10` | InfoReceived | 4 |
| `20` | InTransit | 2 |
| `30` | AvailableForPickup | 2 |
| `40` | DeliveryFailure | 6 |
| `50` | Delivered | 3 |
| `60` | Exception | 6 |

---

### 4.2 查询轨迹接口

**接口路径**：`GET /internal/tracking/:number`

**功能**：查询运单轨迹详情，返回前端规范 JSON。

---

#### 请求参数

| 参数 | 类型 | 位置 | 说明 |
|------|------|------|------|
| `number` | string | URL Path | 运单号 |

---

#### 请求示例

```powershell
curl http://localhost:8082/internal/tracking/YT2606500704802225
```

---

#### 响应示例（成功）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "trackingNumber": "YT2606500704802225",
    "detail": {
      "Item": {
        "TrackingNumber": "USPS123456789",
        "WayBillNumber": "YT2606500704802225",
        "CarrierName": "云途",
        "ProviderName": "云途物流",
        "TrackingStatus": "20",
        "PackageState": "2",
        "OrderTrackingDetails": [
          {
            "ProcessDate": "2025-05-20T08:00:00Z",
            "ProcessLocation": "深圳仓库",
            "ProcessContent": "快件电子信息已收到",
            "TrackingStatus": "10"
          }
        ]
      }
    },
    "status": 20,
    "serviceClass": "YunExpressService",
    "syncedAt": 1780402275
  }
}
```

---

### 4.3 健康检查接口

**接口路径**：`GET /health`

**功能**：检查服务健康状态。

---

#### 响应示例

```json
{
  "code": 200,
  "message": "healthy",
  "data": {
    "service": "yunexpress-webhook",
    "status": "ok",
    "version": "1.0.0"
  }
}
```

---

## 五、补充说明

### 5.1 开发环境 vs 生产环境

| 配置项 | 开发环境 | 生产环境 |
|--------|---------|---------|
| **签名验证** | `YunExpressSkipSignature: true` | `false`（必须验证） |
| **服务发现** | 直连 `127.0.0.1:50051` | Consul `127.0.0.1:8500` |
| **数据库连接** | `sslmode=disable` | 启用 SSL |
| **日志级别** | `info` | `warn` 或 `error` |
| **端口** | `8082`（HTTP） + `50051`（gRPC） | 使用环境变量 |

---

### 5.2 Asynq 异步队列设计（规划中）

**用途**：处理耗时任务（批量查询、数据同步、失败重试）。

**架构**：
```
Webhook 接收 → Redis Stream → Asynq Worker → 批量入库
```

**任务类型**：
- `task:sync_tracking`：批量同步运单轨迹
- `task:retry_failed`：失败重试（最多 3 次）
- `task:cleanup_expired`：清理过期缓存数据

---

### 5.3 云途 Webhook 配置指南

#### 步骤 1：映射公网地址（cpolar）

```powershell
# 安装 cpolar（国内内网穿透工具）
# 访问：https://www.cpolar.com/

# 登录管理后台：http://localhost:9200

# 创建 HTTP 隧道（本地端口 8082）
# 复制公网地址：https://abc123.cpolar.com
```

---

#### 步骤 2：配置云途开发者平台

1. **登录云途开发者平台**
2. **进入**：开发者配置 → Webhook 设置
3. **配置 URL**：
   - Webhook URL：`https://abc123.cpolar.com/webhook/yunexpress/tracking/normal`
   - Secret：`test-secret`（与配置文件一致）
   - 事件类型：勾选「轨迹更新事件」
4. **保存配置**
5. **发送测试报文**

---

#### 步骤 3：验证接收成功

**Go 日志输出**：
```
{"level":"info","content":"Upsert success: YT2606500704802225, status=20"}
```


十、已知风险与待解决事项

#	风险/问题	          影响	                                当前状态	           建议方案
1	跨境网络方案未确定	 海外云途推送 → 国内服务器，网络延迟和不稳定性	未解决	海外部署 Nginx 反代 + 专线回源，或使用全球加速
2	单实例部署，无高可   服务宕机期间 Webhook 推送丢失	        未解决	多实例部署 + Consul 服务发现 + 负载均衡
3	无监控告警	     生产问题无法及时发现	                        未解决	接入 Prometheus + Grafana，配置 P0 告警规则
4	无数据归档策略	     tracking_details 随时间膨胀	                未解决	定期归档已完成运单（如 Delivered 超 90 天）
5	生产环境配置管理	 YAML 直连不适合多环境	                    未解决	Viper 远程配置 / Consul KV / 环境变量
6	限流数值未明确	     不知道服务能扛多少 QPS	                        未解决	压测确定瓶颈，配置合理限流阈值

