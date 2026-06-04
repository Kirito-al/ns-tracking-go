# YunExpress Webhook Service - 重构实施计划

> **项目目标**：完成云途物流 Webhook 改造（Pull → Push），实现 Go标准化接收、Ruby读缓存、最小风险迁移
> **当前状态**：代码审查问题全部修复、Go Workspaces架构改造完成、核心逻辑已实现
> **实施原则**：最小改动、最小风险、分阶段灰度

---

## 一、技术方案概览（Pull → Push 改造）

### 1.1 架构改造方向

**改造目标**：
```
Pull模式（Ruby定时轮询） → Push模式（云途主动推送） + Cache（Ruby读缓存）
```

**核心改造**：
- **Go侧**：标准化接收云途Webhook推送，验签、格式化、幂等入库
- **Ruby侧**：保留现有查询出口和业务后处理，从缓存读取detail数据
- **第一阶段**：保留Pull兜底（YUN_EXPRESS_FALLBACK_PULL开关），避免推送丢失

### 1.2 关键数据契约

**detail JSONB 结构要求**（对标Ruby tracking_detail.rb:282-288）：
```json
{
  "response": {
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
          "TrackingStatus": "10",
          "TrackNodeCode": "XD/ZC",
          "TrackCodeDescription": "揽收",
          "ProcessTimezone": "Asia/Shanghai",
          "Node": null
        }
      ]
    }
  },
  "synced_at": "2026-06-02T10:00:00Z"
}
```

**缓存命中条件**（对标Ruby tracking_detail.rb:246-252）：
1. `detail['response']['Item']` 存在（完整结构）
2. `synced_at` 在 `CACHE_HOURS`（4小时）内
3. 任一条件不满足 → Ruby走Pull兜底

---

## 二、当前实现状态（已实现逻辑）

### 2.1 已实现核心逻辑

| 模块 | 实现文件 | 对标Ruby | 状态 |
|------|---------|----------|------|
| **YunExpressFormatter** | `service/tracking/api/internal/formatter/yunexpress_formatter.go` | `yun_express_track_formatter.rb` | ✅ 已实现 |
| **状态码映射** | `service/tracking/api/internal/formatter/status_codes.go` | `STATUS_CODES` 常量 | ✅ 已实现 |
| **PackageState计算** | `status_codes.go:GetPackageState()` | `STATUS_CODES_TO_PACKAGE_STATES` | ✅ 已实现 |
| **TIS Push转换** | `service/tracking/api/internal/logic/webhook/tis_push_logic.go` | 云途Webhook payload解析 | ✅ 已实现 |
| **Upsert四表同步** | `service/tracking/rpc/internal/logic/upsert_logic.go` | `tracking_detail.rb:sync!` | ✅ 已实现 |
| **幂等去重** | `upsert_logic.go:44-60` | Redis幂等key | ✅ 已实现 |
| **海外包裹判断** | `upsert_logic.go:mapTrackStatus()` | `overseas_package_tracking_log.rb` | ✅ 已实现 |

### 2.2 已实现数据库表结构

| 表名 | Model文件 | 关键字段 | 对标Ruby | 状态 |
|------|----------|---------|----------|------|
| **tracking_details** | `models.go:TrackingDetail` | detail (JSONB), synced_at | `tracking_detail.rb` | ✅ 已定义 |
| **tracking_logs** | `models.go:TrackingLog` | source_tracking_number, channel_alias, 5个时间戳 | `tracking_log.rb` | ✅ 已定义 |
| **tracking_cache_log** | `models.go:TrackingCacheLog` | detail (JSONB), synced_at | `tracking_cache_log.rb` | ✅ 已定义 |
| **overseas_package_tracking_logs** | `overseas_package_model.go` | tracking_number, overseas_package | `overseas_package_tracking_log.rb` | ✅ 已定义 |

### 2.3 已实现状态码映射

**STATUS_CODES**（12项完整映射）：
```go
"0":    "NotFound",
"10":   "InfoReceived",
"20":   "InTransit",
"30":   "AvailableForPickup",
"40":   "DeliveryFailure",
"50":   "Delivered",
"60":   "Exception",
"70":   "Expired",
"80":   "Exception",  // 海关查验（会被改写为20）
"90":   "Exception_Returned",
"100":  "Exception_Cancel",
"1001": "InTransit_Arrival"
```

**STATUS_CODES_TO_PACKAGE_STATES**（11项映射）：
```go
"0":    "0",  // Undefined
"10":   "4",  // Received
"20":   "2",  // InTransit
"30":   "2",  // AvailableForPickup → InTransit
"40":   "6",  // DeliveryFailure
"50":   "3",  // Delivered
"60":   "6",  // Exception
"70":   "6",  // Expired
"80":   "6",  // Exception（改写前）
"90":   "7",  // Exception_Returned → Returned
"100":  "5",  // Exception_Cancel → Canceled
```

**PackageState 文本映射**：
```go
"0": "Undefined",
"1": "Submitted",
"2": "In Transit",
"3": "Delivered",
"4": "Received",
"5": "Canceled",
"6": "Delivery Failed",
"7": "Returned"
```

---

## 三、待补充实现逻辑（基于三份技术方案）

### 3.1 Webhook Payload 结构确认

**当前实现**：已实现 TIS Push 转换逻辑（`tis_push_logic.go:convertTisToWebhook()`）

**待补充**：
- **真实Payload样本**：需从云途沙箱获取 tisPushData 真实结构（验证字段是否与文档一致）
- **三种信封格式兼容**：
  1. 标准格式：`{"data": {"tisPushData": {...}}}`
  2. 直接body：`{"tisPushData": {...}}`
  3. AES加密：`{"encrypted_data": "..."}`（需解密）

**实施步骤**：
1. 联系云途获取沙箱环境测试账号
2. 配置Webhook接收URL（使用cpolar映射本地8082端口）
3. 发送测试推送，记录真实Payload结构
4. 更新 `types.go:TisPushData` 结构体定义（如有差异）

### 3.2 Webhook 签名校验实现

**当前实现**：无（开发环境跳过验签 `YunExpressSkipSignature: true`）

**待补充**（对标tms-talk实现）：
- **签名算法**：SHA-256（非HMAC）
- **签名计算**：`SHA256(timestamp + encrypt_key + raw_body)`
- **Header字段**：`X-YunExpress-Signature`, `X-YunExpress-Timestamp`, `X-YunExpress-Encrypt-Key`

**实施步骤**：
1. 在 `handler/webhook/webhook_handler.go` 添加签名校验中间件
2. 实现 SHA-256 签名计算函数（`utils/signer.go`）
3. 对比 Ruby tms-talk 实现（`D:\gohome\shenhe\ns-admin`）
4. 添加配置开关 `YunExpressSkipSignature`（开发环境）

**代码位置**：
```
service/tracking/api/internal/utils/signer.go  (新增签名校验逻辑)
service/tracking/api/internal/handler/webhook/webhook_handler.go  (调用签名校验)
```

### 3.3 AES-CBC Payload 解密实现

**当前实现**：无

**待补充**（对标tms-talk AES解密）：
- **加密场景**：云途推送加密Payload（`{"encrypted_data": "..."}`）
- **解密算法**：AES-256-CBC + PKCS7Padding
- **密钥来源**：云途开发者平台申请（与签名密钥不同）

**实施步骤**：
1. 在 `utils/crypto.go` 实现 AES-CBC 解密
2. 在 `tis_push_logic.go` 添加加密Payload判断逻辑
3. 对比 Ruby tms-talk 实现（查找 AES 解密代码）
4. 添加配置 `YunExpressEncryptKey`（生产密钥）

**代码位置**：
```
service/tracking/api/internal/utils/crypto.go  (新增AES解密)
service/tracking/api/internal/logic/webhook/tis_push_logic.go  (调用解密)
```

### 3.4 节点代码 → TrackingStatus 完整映射表

**当前实现**：部分映射（`normalize/node_code_mapper.go`）

**待补充**（完整节点代码映射）：
- **来源**：云途节点代码表（需从云途文档或Ruby代码获取）
- **关键节点**：
  - `XD/ZC` → 10（揽收）
  - `XD/CS` → 10（揽收）
  - `FIRST_MILE_ARRIVE` → 10（揽收）
  - `INTERNATIONAL_SHIP` → 20（国际运输）
  - `CUSTOMS_CLEARANCE` → 20（海关清关）
  - `LAST_MILE_DELIVERY` → 30（尾程派送）
  - `DELIVERED` → 50（签收）
  - `MD/TT` → 50（签收）

**实施步骤**：
1. 读取 Ruby `yun_express_track_formatter.rb` 完整节点映射表
2. 补充 `normalize/node_code_mapper.go` 缺失映射
3. 添加映射测试（`formatter_test.go`）

**代码位置**：
```
service/tracking/api/internal/normalize/node_code_mapper.go  (补充映射表)
service/tracking/api/internal/formatter/formatter_test.go  (添加测试)
```

### 3.5 降级策略与监控告警

**当前实现**：无

**待补充**：
- **Pull兜底开关**：`YUN_EXPRESS_FALLBACK_PULL`（Ruby环境变量）
- **监控指标**：
  - Webhook接收成功率（200/500比例）
  - 签名校验失败率
  - 幂等去重命中率
  - 数据库Upsert耗时
- **告警规则**：
  - P0：Webhook接收失败率 > 5%
  - P1：签名校验失败率 > 10%
  - P2：数据库Upsert耗时 > 500ms

**实施步骤**：
1. 在 Ruby 添加 Pull兜底逻辑（`tracking_detail.rb:fetch_cache`）
2. 在 Go 添加 Prometheus 指标埋点（`middleware/metrics.go`）
3. 配置 Grafana 告警规则（对接Prometheus）

**代码位置**：
```
Ruby: ns-tracking/lib/tracking_detail.rb (添加Pull兜底逻辑)
Go: service/tracking/api/internal/middleware/metrics.go (新增埋点)
```

---

## 四、分阶段实施计划（8个阶段）

### Phase 1：Go标准化接收 + 缓存写入（第1-2周）

**目标**：完成Go侧标准化接收逻辑，确保 detail JSONB 结构符合Ruby缓存语义

**实施步骤**：
1. **Step 1.1**：验证 YunExpressFormatter 输出结构
   - 确认 `response.Item` 结构完整（包含TrackingNumber、WayBillNumber、TrackingStatus、OrderTrackingDetails）
   - 确认 `synced_at` 为ISO8601格式（`time.Now().UTC().Format(time.RFC3339)`）
   - 单元测试验证（`formatter_test.go`）

2. **Step 1.2**：验证 Upsert 四表同步逻辑
   - 确认 tracking_details 写入（detail JSONB + synced_at）
   - 确认 tracking_logs 同步（track_status + 5个时间戳）
   - 确认 overseas_package 判断逻辑正确
   - 集成测试验证（使用真实Payload样本）

3. **Step 1.3**：Ruby缓存命中规则复刻
   - 对标 `tracking_detail.rb:fetch_cache` 逻辑
   - 检查 `detail['response']['Item']` 存在
   - 检查 `synced_at` 在4小时内（CACHE_HOURS = 4）
   - 添加Ruby单元测试验证缓存命中

**验收标准**：
- Go Upsert 写入 tracking_details，detail 结构包含完整 response.Item
- Ruby 从缓存读取，命中条件满足时不走Pull
- 单元测试覆盖率 > 80%

**风险点**：
- detail 结构不完整 → Ruby不命中缓存 → 走Pull兜底（需验证）

---

### Phase 2：Webhook签名校验 + AES解密（第2-3周）

**目标**：完成生产级安全校验，防止恶意推送和报文篡改

**实施步骤**：
1. **Step 2.1**：Webhook签名校验实现
   - 实现 SHA-256 签名计算（`utils/signer.go`）
   - 添加 Handler 中间件（验签拦截）
   - 对比 Ruby tms-talk 实现（确保算法一致）
   - 添加签名校验测试（伪造签名拦截）

2. **Step 2.2**：AES-CBC Payload 解密实现
   - 实现 AES-256-CBC 解密（`utils/crypto.go`）
   - 添加加密Payload判断逻辑（`{"encrypted_data": "..."}`）
   - 对比 Ruby tms-talk AES 解密代码
   - 添加解密测试（加密Payload解析）

3. **Step 2.3**：三种信封格式兼容
   - 标准格式：`{"data": {"tisPushData": {...}}}`
   - 直接body：`{"tisPushData": {...}}`
   - AES加密：`{"encrypted_data": "..."}`（解密后解析）
   - 添加兼容性测试（三种格式解析）

**验收标准**：
- 签名校验拦截率 100%（伪造签名）
- AES解密成功率 100%（加密Payload）
- 三种信封格式兼容解析

**风险点**：
- 签名算法与云途不一致 → 推送被拦截（需沙箱验证）
- AES密钥与云途不一致 → 解密失败（需云途申请密钥）

---

### Phase 3：节点代码映射 + 状态码标准化（第3-4周）

**目标**：完成完整节点代码映射表，确保状态码转换符合Ruby语义

**实施步骤**：
1. **Step 3.1**：补充节点代码映射表
   - 读取 Ruby `yun_express_track_formatter.rb` 完整映射
   - 补充 `normalize/node_code_mapper.go` 缺失节点
   - 添加映射测试（关键节点验证）

2. **Step 3.2**：海关查验状态改写验证
   - 状态80（海关查验）→ 改写为20（InTransit）
   - 对标 Ruby `customs_fixer.rb` 逻辑
   - 确认 `FixCustomsInspectionStatusInEvents()` 实现正确

3. **Step 3.3**：PackageState 计算验证
   - 状态码 → PackageState 映射（11项）
   - 对标 Ruby `STATUS_CODES_TO_PACKAGE_STATES` 常量
   - 添加 PackageState 测试（状态码映射验证）

**验收标准**：
- 节点代码映射覆盖率 > 95%（关键节点全覆盖）
- 海关查验状态改写正确（80 → 20）
- PackageState 计算正确（状态码映射）

**风险点**：
- 节点代码映射缺失 → TrackingStatus 错误 → 缓存不命中

---

### Phase 4：降级策略 + 监控告警（第4-5周）

**目标**：完成生产级降级策略和监控告警，确保灰度发布可控

**实施步骤**：
1. **Step 4.1**：Pull兜底开关配置
   - Ruby环境变量：`YUN_EXPRESS_FALLBACK_PULL=true`（第一阶段启用）
   - 添加开关逻辑（缓存不命中 → Pull兜底）
   - 灰度验证（部分运单走Push，部分走Pull）

2. **Step 4.2**：监控指标埋点
   - Webhook接收成功率（Prometheus Counter）
   - 签名校验失败率（Prometheus Counter）
   - 幂等去重命中率（Redis监控）
   - 数据库Upsert耗时（Prometheus Histogram）

3. **Step 4.3**：Grafana告警配置
   - P0：Webhook接收失败率 > 5%（钉钉告警）
   - P1：签名校验失败率 > 10%（邮件告警）
   - P2：数据库Upsert耗时 > 500ms（日志告警）

**验收标准**：
- Pull兜底开关可配置（Ruby环境变量）
- Prometheus 指标采集正常（Grafana 展示）
- 告警规则触发正常（测试告警）

**风险点**：
- 监控缺失 → 生产问题无法及时发现

---

### Phase 5：灰度发布 + 流量切换（第5-6周）

**目标**：完成灰度发布验证，逐步切换流量到Webhook模式

**实施步骤**：
1. **Step 5.1**：灰度发布配置
   - Go侧灰度开关：`GRAY_SCALE_ENABLED=true`（配置文件）
   - Ruby侧灰度开关：`YUN_EXPRESS_WEBHOOK_ENABLED=true`（环境变量）
   - 灰度比例：10% → 30% → 50% → 100%

2. **Step 5.2**：流量切换验证
   - 10%流量走Webhook，90%走Pull（观察1周）
   - 30%流量走Webhook（观察1周）
   - 50%流量走Webhook（观察1周）
   - 100%流量走Webhook（关闭Pull）

3. **Step 5.3**：数据一致性验证
   - 对比 Go缓存数据 vs Ruby Pull数据（一致性检查）
   - 状态码映射一致性验证
   - PackageState 一致性验证

**验收标准**：
- 灰度发布可控（流量切换平滑）
- 数据一致性 100%（缓存 vs Pull）
- 无告警触发（灰度期间）

**风险点**：
- 流量切换过快 → 异常无法及时发现

---

### Phase 6：Pull兜底关闭 + 生产验证（第6-7周）

**目标**：关闭Pull兜底，完成Webhook模式生产验证

**实施步骤**：
1. **Step 6.1**：关闭Pull兜底开关
   - Ruby环境变量：`YUN_EXPRESS_FALLBACK_PULL=false`
   - 观察告警指标（无异常）
   - 监控缓存命中率（> 95%）

2. **Step 6.2**：生产验证
   - 真实运单推送验证（全量运单）
   - 数据库写入验证（tracking_details 数据完整）
   - Ruby查询验证（缓存命中正常）

3. **Step 6.3**：性能验证
   - Webhook接收耗时 < 200ms
   - Upsert耗时 < 500ms
   - Ruby缓存查询耗时 < 50ms

**验收标准**：
- Pull兜底关闭（无Pull调用）
- 缓存命中率 > 95%
- Webhook接收成功率 > 99%

**风险点**：
- 推送丢失 → 数据缺失（需监控告警）

---

### Phase 7：监控优化 + 压力测试（第7-8周）

**目标**：完成生产级监控优化和压力测试验证

**实施步骤**：
1. **Step 7.1**：监控优化
   - 补充业务指标（签收率、异常率、平均运输天数）
   - Grafana Dashboard 完善（业务监控面板）
   - 告警规则优化（P0/P1/P2分级）

2. **Step 7.2**：压力测试
   - 模拟高并发Webhook推送（1000 QPS）
   - 验证幂等去重性能（Redis吞吐）
   - 验证数据库Upsert性能（PostgreSQL JSONB索引）

3. **Step 7.3**：容量规划
   - Webhook服务实例数（根据QPS峰值）
   - Redis容量（幂等key存储）
   - PostgreSQL容量（tracking_details表增长）

**验收标准**：
- 压力测试通过（1000 QPS无异常）
- 监控指标完整（业务 + 技术）
- 容量规划清晰（未来6个月）

**风险点**：
- 高并发推送 → 服务雪崩（需限流）

---

### Phase 8：文档完善 + 团队培训（第8周）

**目标**：完成技术文档和团队培训，确保后续维护可控

**实施步骤**：
1. **Step 8.1**：技术文档完善
   - Webhook接口文档（对接云途）
   - Go服务运维文档（部署、监控、告警）
   - Ruby缓存逻辑文档（缓存命中规则）

2. **Step 8.2**：团队培训
   - Go服务运维培训（运维团队）
   - Ruby缓存逻辑培训（开发团队）
   - 告警处理培训（值班团队）

3. **Step 8.3**：应急预案
   - Webhook服务宕机 → Pull兜底开关切换
   - 数据库故障 → Asynq重试机制
   - Redis故障 → 幂等去重降级

**验收标准**：
- 技术文档完整（运维 + 开发）
- 团队培训完成（运维 + 开发 + 值班）
- 应急预案清晰（故障处理流程）

---

## 五、关键风险点与缓解措施

### 5.1 技术风险

| 风险点 | 影响 | 缓解措施 |
|--------|------|---------|
| **detail结构不完整** | Ruby不命中缓存，走Pull兜底 | 单元测试验证response.Item结构完整性 |
| **签名算法不一致** | Webhook推送被拦截 | 沙箱验证签名算法，对比Ruby tms-talk实现 |
| **节点代码映射缺失** | TrackingStatus错误，缓存不命中 | 补充完整映射表，单元测试验证关键节点 |
| **幂等key过期时间** | 重复推送导致数据覆盖 | Redis key过期时间1小时，确保幂等窗口 |
| **海外包裹判断缺失** | InfoReceived状态映射错误 | 验证overseas_package判断逻辑 |

### 5.2 业务风险

| 风险点 | 影响 | 缓解措施 |
|--------|------|---------|
| **推送丢失** | 运单轨迹缺失 | Pull兜底开关，监控告警触发 |
| **流量切换过快** | 异常无法及时发现 | 灰度发布10%→30%→50%→100% |
| **缓存过期** | Ruby查询空数据 | CACHE_HOURS=4，定期刷新缓存 |
| **跨境网络不稳定** | Webhook推送延迟 | 海外部署Nginx反代 + 专线回源 |

### 5.3 运维风险

| 风险点 | 影响 | 缓解措施 |
|--------|------|---------|
| **单实例部署** | 服务宕机期间推送丢失 | 多实例部署 + Consul服务发现 |
| **无监控告警** | 生产问题无法及时发现 | Prometheus + Grafana告警 |
| **数据归档缺失** | tracking_details表膨胀 | 定期归档已完成运单（超90天） |

---

## 六、验收标准（灰度准入条件）

### 6.1 Go服务验收标准

| 验收项 | 标准 | 验证方式 |
|--------|------|---------|
| **编译通过** | 无编译错误 | `make build` 成功 |
| **单元测试** | 覆盖率 > 80% | `go test -cover ./...` |
| **Upsert逻辑** | detail JSONB结构完整 | 单元测试验证response.Item |
| **幂等去重** | 重复推送拦截 | Redis幂等key验证 |
| **签名校验** | 伪造签名拦截 | 签名校验测试 |
| **AES解密** | 加密Payload解析 | 解密测试 |

### 6.2 Ruby缓存验收标准

| 验收项 | 标准 | 验证方式 |
|--------|------|---------|
| **缓存命中** | 命中率 > 95% | 监控统计 |
| **detail结构** | response.Item完整 | 单元测试验证 |
| **synced_at检查** | 4小时窗口正确 | 时间戳验证 |
| **Pull兜底** | 开关可配置 | 环境变量验证 |

### 6.3 监控告警验收标准

| 验收项 | 标准 | 验证方式 |
|--------|------|---------|
| **Prometheus指标** | 采集正常 | Grafana面板展示 |
| **P0告警** | 触发钉钉告警 | 模拟告警测试 |
| **P1告警** | 触发邮件告警 | 模拟告警测试 |
| **P2告警** | 触发日志告警 | 模拟告警测试 |

---

## 七、附录：关键代码位置参考

### 7.1 Go服务关键文件

| 文件路径 | 功能 | 对标Ruby |
|---------|------|----------|
| `service/tracking/api/internal/formatter/yunexpress_formatter.go` | YunExpress格式化器 | `yun_express_track_formatter.rb` |
| `service/tracking/api/internal/formatter/status_codes.go` | 状态码映射 | `STATUS_CODES` 常量 |
| `service/tracking/api/internal/normalize/node_code_mapper.go` | 节点代码映射 | 云途节点代码表 |
| `service/tracking/api/internal/logic/webhook/tis_push_logic.go` | TIS Push转换 | 云途Webhook payload解析 |
| `service/tracking/rpc/internal/logic/upsert_logic.go` | Upsert四表同步 | `tracking_detail.rb:sync!` |
| `service/tracking/rpc/internal/model/models.go` | 数据库Model | `tracking_detail.rb` |
| `service/tracking/rpc/internal/dao/tracking_dao.go` | Upsert DAO | 数据库Upsert实现 |

### 7.2 Ruby服务关键文件（参考）

| 文件路径 | 功能 | 用途 |
|---------|------|------|
| `D:\gohome\shenhe\ns-tracking\lib\tracking_detail.rb` | 轨迹详情Model | 缓存命中规则、Pull兜底逻辑 |
| `D:\gohome\shenhe\ns-tracking\lib\yun_express_track_formatter.rb` | YunExpress格式化器 | 状态码映射、节点代码映射 |
| `D:\gohome\shenhe\ns-tracking\lib\tracking_cache_log.rb` | 缓存日志Model | 缓存写入逻辑 |
| `D:\gohome\shenhe\ns-admin\lib\tms-talk.rb` | TMS接口封装 | Webhook签名校验、AES解密 |

---

## 八、总结：Webhook改造核心要点

### 8.1 核心改造原则

1. **最小改动**：Go只做标准化接收，Ruby保留后处理
2. **最小风险**：保留Pull兜底，灰度发布验证
3. **数据契约**：detail JSONB结构必须包含完整response.Item
4. **幂等设计**：Redis幂等key（upsert:{tracking_number}:{synced_at}），1小时过期
5. **降级策略**：YUN_EXPRESS_FALLBACK_PULL开关，缓存不命中走Pull

### 8.2 关键技术要点

1. **状态码映射**：
   - Webhook package_status: N/F/T/D/E/R/C → TrackingStatus: 0/10/20/50/40/90/100
   - TrackingStatus → PackageState: 11项映射（STATUS_CODES_TO_PACKAGE_STATES）

2. **海关查验改写**：
   - 状态80（海关查验）→ 改写为20（InTransit）
   - 对标Ruby customs_fixer.rb逻辑

3. **海外包裹判断**：
   - InfoReceived + overseas_package → track_status=1（in_transit）
   - InfoReceived + NO overseas_package → track_status=0（to_receive）

4. **节点代码映射**：
   - XD/ZC → 10（揽收）
   - INTERNATIONAL_SHIP → 20（国际运输）
   - DELIVERED/MD/TT → 50（签收）

### 8.3 后续维护建议

1. **定期验证**：每月检查缓存命中率、Webhook接收成功率
2. **容量规划**：每季度评估tracking_details表增长，制定归档策略
3. **监控优化**：根据业务需求补充监控指标（签收率、异常率）
4. **应急预案**：定期演练故障处理流程（Webhook宕机 → Pull切换）

---

**文档版本**：v1.0
**创建时间**：2026-06-04
**负责人**：Go团队 + Ruby团队
**状态**：待实施（Phase 1启动）