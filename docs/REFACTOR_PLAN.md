# ns-tracking Ruby → Go 重构方案

> 项目重构：从 Ruby on Rails 迁移到 Go-Zero 微服务架构
> 目标：Webhook Push 模式 + 事件驱动 + 异步队列

---

## 一、项目背景

### 1.1 现有 Ruby 架构

**源码地址**：
- `D:\gohome\shenhe\ns-admin`（Ruby 后台管理）
- `D:\gohome\shenhe1\ns-tracking`（Ruby 物流轨迹服务）

**核心流程**（Pull 模式）：

```
客户端请求 → TrackingController
              ↓
         TrackingService（路由分发）
              ↓
         YunExpressService.new(tracking_number:).call_and_cache()
              ↓
         HTTP GET → 云途 API (oms.api.yunexpress.com)
              ↓（实时拉取，同步等待）
         YunExpressTrackFormatter（格式化）
              ↓
         BaseExpressService#adjust（后处理）
              ↓
         返回 JSON → 客户端
```

**痛点**：
- ❌ 每次查询都调用云途 API（QPS 高时压力大）
- ❌ 响应延迟 1~10s（依赖云途 API 性能）
- ❌ 云途 API 故障 → 查询全部失败（无降级）
- ❌ 无法处理高频推送场景

---

### 1.2 目标 Go 架构

**核心变化**：Pull → Push（Webhook）

```
云途 Webhook 推送 → Go Webhook Service
                      ↓（异步接收）
                   签名验证 + 格式化
                      ↓
                   写入 PostgreSQL（JSONB）
                      ↓
客户端查询 → Ruby TrackingController
              ↓（读缓存）
         TrackingDetail DB（<10ms）
              ↓
         返回 JSON → 客户端
```

**优势**：
- ✅ 查询响应 <10ms（读 DB 缓存）
- ✅ 云途 API 调用量降低 90%+
- ✅ 云途推送失败不影响查询（降级到上次缓存）
- ✅ 支持实时推送（状态变更立即入库）

---

## 二、当前进度（Phase 1 已完成）

### 2.1 已完成模块

| 模块 | 状态 | 说明 |
|------|------|------|
| Webhook 接收 | ✅ | `POST /webhook/yunexpress/tracking/normal` |
| 事件驱动 | ✅ | `TrackingUpsertedEvent` + `LogListener` |
| Asynq 队列 | ✅ | Worker 进程 + 重试机制 |
| JSONB 存储 | ✅ | `detail` 字段改为 `jsonb` 类型 |
| 数据库迁移 | ✅ | `database_migration_phase1.sql` |
| 全链路测试 | ✅ | 云途推送验证通过 |

### 2.2 当前 Go 项目结构

```
D:\gohome\demo1-gozero\ns-tracking-go
├── tracking-api/           # HTTP API 服务
│   ├── internal/handler/   # Webhook Handler
│   ├── internal/logic/     # 业务逻辑
│   └ tracking-api.go       # 服务入口
│
├── tracking-srv/           # gRPC 服务
│   ├── internal/dao/       # 数据访问层
│   ├── internal/event/     # 事件系统
│   ├── internal/queue/     # Asynq 队列
│   ├── cmd/worker/         # Worker 进程
│   └ tracking.go           # gRPC 服务入口
│
├── common/                 # 公共模块
├── database.sql            # 表结构
└── docker-compose.yml      # 中间件配置
```

---

## 三、分天实施计划（共 10 天）

### 📅 Day 1：环境准备 + 项目搭建（已完成）

**目标**：Go 项目骨架创建

**任务清单**：
- [x] 创建 Go-Zero 项目结构（tracking-api + tracking-srv）
- [x] 配置 docker-compose（PostgreSQL + Redis）
- [x] 编写 database.sql（四张核心表）
- [x] 创建 .gitignore（排除敏感文件）
- [x] 验证数据库连接

**关键文件**：
- `database.sql`（tracking_details/tracking_logs/tracking_cache_logs/tracking_replaces）
- `docker-compose.yml`（中间件配置）

---

### 📅 Day 2：Webhook 接收 + HMAC 签名验证（已完成）

**目标**：接收云途 Webhook 推送

**任务清单**：
- [x] 创建 Webhook Handler（`POST /webhook/yunexpress/tracking/normal`）
- [x] 实现 HMAC-SHA256 签名验证
- [x] 解析云途 Payload（JSON 结构）
- [x] 返回 200 OK（即使入库失败也返回成功）

**关键代码**：
```go
// tracking-api/internal/handler/webhook/webhook_handler.go
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. 验签
    if !h.verifySignature(r) {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    
    // 2. 解析 Payload
    var payload YunExpressWebhookPayload
    json.NewDecoder(r.Body).Decode(&payload)
    
    // 3. gRPC 调用入库
    h.grpcClient.UpsertTracking(...)
    
    w.WriteHeader(http.StatusOK)
}
```

**测试验证**：
- APIPOST 推送测试数据
- 终端日志：`Upsert success: YT2606500704802225, status=20`

---

### 📅 Day 3：事件驱动模型 + Asynq 队列（已完成）

**目标**：入库成功后触发事件 + 异步重试

**任务清单**：
- [x] 创建 Event 系统（define/dispatcher/listener）
- [x] 实现 `TrackingUpsertedEvent`（成功事件）
- [x] 实现 `LogListener`（日志打印）
- [x] 创建 Asynq Worker 进程
- [x] 实现失败重试（MaxRetry 3）

**关键代码**：
```go
// tracking-srv/internal/logic/upsert_logic.go
if l.svcCtx.EventDispatcher != nil {
    event := define.NewTrackingUpsertedEvent(...)
    l.svcCtx.EventDispatcher.Dispatch(event)
}

if l.svcCtx.AsynqClient != nil {
    task := asynq.NewTask(tasks.TaskRetryFailed, payload)
    l.svcCtx.AsynqClient.Enqueue(task)
}
```

**启动流程**：
```powershell
# Worker 进程
go run tracking-srv/cmd/worker/main.go
```

---

### 📅 Day 4：数据格式化 + 状态码映射（对标 Ruby）

**目标**：完全复刻 Ruby YunExpressTrackFormatter

**任务清单**：
- [ ] 实现 `TrackFormatter`（格式化器）
- [ ] 状态码映射（STATUS_CODES + STATUS_CODES_TO_PACKAGE_STATES）
- [ ] PackageState 计算（对标 Ruby）
- [ ] 轨迹事件排序（sort_by ProcessDate）
- [ ] 事件过滤（TrackingEventFilter）

**Ruby 对照**：
```ruby
# YunExpressTrackFormatter::STATUS_CODES
{
  '0' => 'NotFound',
  '10' => 'InfoReceived',
  '20' => 'InTransit',
  '30' => 'AvailableForPickup',
  '40' => 'DeliveryFailure',
  '50' => 'Delivered',
  '60' => 'Exception',
  '70' => 'Expired',
  '80' => 'Exception',
  '90' => 'Exception_Returned',
  '100' => 'Exception_Cancel'
}
```

**Go 实现**：
```go
// tracking-srv/internal/formatter/status_codes.go
var StatusCodes = map[string]string{
    "0":   "NotFound",
    "10":  "InfoReceived",
    "20":  "InTransit",
    // ...完全对标 Ruby
}

var StatusCodesToPackageStates = map[string]string{
    "0":   "0", // Undefined
    "10":  "4", // Received
    "20":  "2", // InTransit
    "50":  "3", // Delivered
    // ...完全对标 Ruby
}
```

---

### 📅 Day 5：后处理管道（地点脱敏 + 头程拼接）

**目标**：复刻 Ruby BaseExpressService#adjust

**任务清单**：
- [ ] 地点脱敏（`TrackLocationMasking`）
- [ ] 头程轨迹拼接（`append_pack_tracking`）
- [ ] 替换假轨迹（`replace_with_details`）
- [ ] 移除中国省份信息（`remove_chinese_provinces_from_content!`）
- [ ] 日期调整（`adjust_tracking_date`）

**Ruby 对照**：
```ruby
# base_express_service.rb:33-78
def adjust(resp)
  resp = replace_delivered_to_with_arrived_at(resp)
  resp = append_pack_tracking(resp, tracking_log)
  resp = tracking_detail.replace_with_details(resp)
  resp = TrackingEventFilter.filter_order_tracking_details(...)
  mask_process_location_by_destination!(resp, tracking_log: tracking_log)
  remove_chinese_provinces_from_content!(resp)
  resp
end
```

**关键逻辑**：
- **地点脱敏**：只保留目的国的 location（如 NL），其他国家/城市置空
- **头程拼接**：从 OverseasPackage 获取头程轨迹并 prepend
- **事件过滤**：移除重复/无效事件

---

### 📅 Day 6：Token 路由 + GE 渠道判断

**目标**：处理普通渠道 vs GE-云途渠道

**任务清单**：
- [ ] 实现 Token Resolver（Token 路由）
- [ ] GE 渠道正则判断（`GE-云途 / GE云途 / GE-YT / GEYT`）
- [ ] 读取 tracking_logs 判断渠道
- [ ] 环境变量配置（YUN_EXPRESS_TOKEN / GE_YUN_EXPRESS_TOKEN）

**Ruby 对照**：
```ruby
# yun_express_service.rb
def authorization_token
  if is_ge_yun_channel?
    ENV['GE_YUN_EXPRESS_TOKEN']
  else
    ENV['YUN_EXPRESS_TOKEN']
  end
end

def is_ge_yun_channel?
  tracking_log = TrackingLog.find_by(source_tracking_number: tracking_number)
  pattern = /^(GE-云途|GE云途|GE-YT|GEYT)/i
  tracking_log.channel_alias =~ pattern ||
  tracking_log.shipping_agent =~ pattern
end
```

**Go 实现**：
```go
// tracking-srv/internal/service/token_resolver.go
func ResolveToken(trackingNumber string, db *gorm.DB) string {
    var log TrackingLog
    db.Where("source_tracking_number = ?", trackingNumber).First(&log)
    
    if IsGeYunChannel(log.ChannelAlias, log.ShippingAgent) {
        return os.Getenv("GE_YUN_EXPRESS_TOKEN")
    }
    return os.Getenv("YUN_EXPRESS_TOKEN")
}

func IsGeYunChannel(channelAlias, shippingAgent string) bool {
    pattern := regexp.MustCompile(`(?i)^(GE-云途|GE云途|GE-YT|GEYT)`)
    return pattern.MatchString(channelAlias) || pattern.MatchString(shippingAgent)
}
```

---

### 📅 Day 7：Ruby 侧改造（读缓存优先）

**目标**：修改 Ruby YunExpressService，读 DB 缓存

**任务清单**：
- [ ] 改造 `YunExpressService#call`（读 tracking_details DB）
- [ ] 保留 Pull 降级开关（`YUN_EXPRESS_FALLBACK_PULL`）
- [ ] 编写集成测试（对比新旧两路输出）
- [ ] 验证数据一致性

**Ruby 改造**：
```ruby
# yun_express_service.rb
def call
  # 优先读缓存（Go Webhook 写入）
  detail = TrackingDetail.find_by(tracking_number: tracking_number)
  return detail.detail if detail&.detail&.dig('Item').present?
  
  # 缓存为空时：Pull 降级
  if ENV['YUN_EXPRESS_FALLBACK_PULL'].present?
    pull_from_api  # 原有逻辑
  end
end
```

**数据一致性测试**：
```ruby
# spec/integration/yunexpress_go_ruby_consistency_spec.rb
it 'Go and Ruby output same structure' do
  go_output = GoService.track('YT2606500704802225')
  ruby_output = YunExpressService.track('YT2606500704802225')
  
  expect(go_output['Item']['TrackingStatus']).to eq(ruby_output['Item']['TrackingStatus'])
  expect(go_output['Item']['OrderTrackingDetails'].length).to eq(ruby_output['Item']['OrderTrackingDetails'].length)
end
```

---

### 📅 Day 8：云途 Webhook 配置 + 沙箱测试

**目标**：与云途对接 Webhook 配置

**任务清单**：
- [ ] 配置云途 Webhook URL（`https://your-domain.com/webhook/yunexpress/tracking`）
- [ ] 沙箱环境测试（`openapi.yunexpress.cn`）
- [ ] 确认签名密钥（HMAC-SHA256）
- [ ] 确认推送时机（每次状态变更推送）
- [ ] 确认 Payload 格式（字段定义）

**云途配置清单**：

| 配置项 | 值 | 说明 |
|--------|-----|------|
| Webhook URL | `https://your-domain.com/webhook/yunexpress/tracking` | 公网可达 |
| 签名方式 | HMAC-SHA256 | Header: X-YunExpress-Signature |
| 推送时机 | 每次状态变更 | 实时推送（5~30 分钟） |
| Payload 格式 | JSON | 需对齐 Go 结构体 |

**沙箱测试 API**：
```
https://openapi.yunexpress.cn/v1/track-service/info/get
https://openapi.yunexpress.cn/v1/order/shipping-docs/get
```

---

### 📅 Day 9：灰度上线 + 监控验证

**目标**：先上线 Go 服务，观察数据质量

**任务清单**：
- [ ] Go 服务部署（独立容器）
- [ ] 启动 Worker 进程
- [ ] 云途 Webhook 指向 Go 服务
- [ ] Ruby 保持 Pull 模式（降级保留）
- [ ] 观察 tracking_details 写入情况
- [ ] 监控缓存命中率和延迟

**灰度策略**：
```
Phase A：Go 服务接收 Webhook → 写 DB（Ruby 还在 Pull）
          观察 1 周，确认数据质量

Phase B：切换 Ruby 读缓存优先（Pull 降级保留）
          监控缓存命中率

Phase C：稳定后移除 Pull 降级逻辑
          完全切换到 Webhook Push 模式
```

**监控指标**：
- tracking_details 写入频率
- 缓存命中率（Ruby 读 DB 的成功率）
- 响应延迟（Ruby 读 DB 的 P99）
- Webhook 接收成功率

---

### 📅 Day 10：其他物流商迁移（17Track / SY / YW）

**目标**：复刻云途模式，迁移其他物流商

**任务清单**：
- [ ] SeventeenTrackService 迁移（17Track Webhook）
- [ ] SyTrackFormatter 迁移（SY Track 格式化）
- [ ] YwTrackFormatter 迁移（YW Track 格式化）
- [ ] 多物流商路由逻辑（TrackingService）

**Ruby 物流商清单**：
```
YunExpressService     → 已完成
SeventeenTrackService → Day 10
SyTrackService        → Day 10
YwTrackService        → Day 10
WanguoExpressService  → Day 10（可选）
JxExpressService      → Day 10（可选）
```

**路由逻辑**：
```go
// tracking-api/internal/service/service_resolver.go
func ResolveService(trackingNumber string) string {
    if strings.HasPrefix(trackingNumber, "YT") {
        return "yunexpress"
    }
    if strings.HasPrefix(trackingNumber, "WS") {
        return "yunexpress"  // WS 单号也走云途
    }
    // ...其他物流商判断
    return "17track"  // 默认
}
```

---

## 四、技术栈对比

### 4.1 Ruby vs Go 架构对比

| 模块 | Ruby (现状) | Go (目标) |
|------|------------|-----------|
| **Web 接口** | Rails Controller | Go-Zero REST API |
| **数据接收** | HTTP Pull（实时调用） | Webhook Push（被动接收） |
| **格式化** | YunExpressTrackFormatter | Go TrackFormatter |
| **状态映射** | STATUS_CODES 常量 | Go status_codes.go |
| **事件系统** | 无 | EventDispatcher + Asynq |
| **异步队列** | Sidekiq/SolidQueue | Asynq (Redis) |
| **数据存储** | PostgreSQL (text) | PostgreSQL (JSONB) |
| **日志系统** | Zap + Lumberjack | Zap（终端） |

---

## 五、数据库表结构（已优化）

### 5.1 tracking_details（核心表）

```sql
CREATE TABLE tracking_details (
    id SERIAL PRIMARY KEY,
    tracking_number VARCHAR(255) NOT NULL UNIQUE,
    
    -- JSONB 存储（Phase 1 已完成）
    detail JSONB NOT NULL,        -- 当前轨迹详情
    last_detail JSONB,            -- 上次轨迹备份
    
    -- 状态信息
    status INTEGER,               -- 状态码（0-100）
    service_class VARCHAR(255),   -- 服务类名
    
    -- 时间戳
    synced_at BIGINT,             -- 同步时间（Unix）
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- 索引（Phase 2 可添加）
CREATE INDEX idx_tracking_number ON tracking_details(tracking_number);
CREATE INDEX idx_status ON tracking_details(status);
CREATE INDEX idx_synced_at ON tracking_details(synced_at);
```

### 5.2 tracking_logs（渠道判断表）

```sql
CREATE TABLE tracking_logs (
    id SERIAL PRIMARY KEY,
    source_tracking_number VARCHAR(255),  -- 云途单号
    tracking_number VARCHAR(255),         -- 内部单号
    
    -- 渠道信息（GE 判断）
    channel_alias VARCHAR(255),           -- 渠道别名（GE-云途）
    shipping_agent VARCHAR(255),          -- 物流代理商
    shipping_channel VARCHAR(255),        -- 物流渠道
    
    country_code VARCHAR(50),             -- 目的国
    
    -- 时间戳
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
```

---

## 六、风险点与缓解措施

| 风险 | 等级 | 缓解措施 |
|------|------|---------|
| 云途 Webhook 推送不稳定（漏推） | 中 | 保留定时主动拉取（每小时兜底） |
| Go 写入 JSON 与 Ruby 期望不一致 | 高 | 编写对比测试（同一单号两路对比） |
| GE 渠道无法区分 Token | 中 | 云途配置两个 Webhook 端点 |
| DB 连接竞争（Go & Ruby） | 低 | Go 只写，Ruby 只读 |
| Webhook 签名密钥泄露 | 中 | 存环境变量，定期轮换 |
| 首次上线历史数据空缺 | 中 | 上线前用 Pull 模式预热 |

---

## 七、成功标准

### 7.1 性能指标

| 指标 | 改造前 | 改造后（目标） |
|------|--------|---------------|
| API 查询响应时间 | 1~10s（依赖云途） | <50ms（读 DB） |
| 云途 API 调用量 | 每次查询都调用 | 降低 90%+ |
| 云途 API 故障影响 | 查询全部失败 | 返回最近缓存 |
| 系统可用性 | 依赖外部 API | 提升，外部解耦 |

### 7.2 功能验证

- ✅ 云途推送正常接收（签名验证通过）
- ✅ 数据写入 tracking_details（JSONB 格式正确）
- ✅ Ruby 读取缓存正常（数据结构一致）
- ✅ 降级逻辑验证（缓存 miss 时 Pull 成功）
- ✅ 事件触发正常（LogListener 打印日志）

---

## 八、附录：关键文件对照表

### 8.1 Ruby → Go 文件映射

| Ruby 文件 | Go 文件 | 说明 |
|----------|---------|------|
| `yun_express_service.rb` | `tracking-api/internal/handler/webhook/` | Webhook 接收 |
| `yun_express_track_formatter.rb` | `tracking-srv/internal/formatter/` | 格式化逻辑 |
| `base_express_service.rb` | `tracking-srv/internal/logic/upsert_logic.go` | 后处理管道 |
| `tracking_service.rb` | `tracking-api/internal/service/service_resolver.go` | 路由分发 |
| `tracking_detail.rb` | `tracking-srv/internal/model/models.go` | 数据模型 |
| `tracking_event_filter.rb` | `tracking-srv/internal/filter/` | 事件过滤 |
| `track_location_masking.rb` | `tracking-srv/internal/mask/` | 地点脱敏 |

---

## 九、下一步行动

### 当前状态
- ✅ Phase 1 完成（Day 1-3）
- ⏳ Phase 2 待启动（Day 4-6）
- ⏳ Phase 3 待启动（Day 7-10）

### 建议开始顺序
1. **Day 4**：数据格式化 + 状态码映射（对标 Ruby）
2. **Day 5**：后处理管道（地点脱敏 + 头程拼接）
3. **Day 6**：Token 路由 + GE 渠道判断
4. **Day 7**：Ruby 侧改造（读缓存优先）
5. **Day 8-10**：云途对接 + 灰度上线 + 其他物流商

---

**准备好开始 Day 4 时告诉我，我来帮你实施。**