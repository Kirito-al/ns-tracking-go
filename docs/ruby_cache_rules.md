# Ruby缓存命中规则文档

> 对标：Ruby tracking_detail.rb:246-252 (fetch_cache方法)
> 目的：确保Go写入的detail JSONB结构符合Ruby缓存语义

---

## 一、Ruby缓存命中条件

Ruby从tracking_details表读取detail字段，判断缓存是否命中：

```ruby
# tracking_detail.rb:246-252
def fetch_cache
  # 1. 检查detail['response']['Item']是否存在
  if detail['response'] && detail['response']['Item']
    # 2. 检查synced_at是否在CACHE_HOURS（4小时）内
    if synced_at && synced_at > (Time.now - CACHE_HOURS.hours).to_i
      # 缓存命中 → 返回缓存数据
      return detail['response']['Item']
    end
  end
  
  # 缓存不命中 → 调用Pull兜底
  return nil
end
```

**关键条件**：
1. `detail['response']['Item']` 必须存在（完整结构）
2. `synced_at` 必须在 `CACHE_HOURS`（4小时）内
3. 任一条件不满足 → Ruby调用Pull兜底

---

## 二、Go写入detail结构要求

Go必须写入符合Ruby缓存语义的detail JSONB结构：

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
          "TrackingStatus": "10"
        }
      ]
    }
  },
  "synced_at": "2026-06-04T17:00:00Z"  // ISO8601格式（关键：Ruby解析时间戳）
}
```

**关键要点**：
- `response` 包装层必须存在
- `Item` 结构必须完整（包含TrackingNumber、WayBillNumber、TrackingStatus）
- `synced_at` 必须为ISO8601格式（Ruby解析时间戳）

---

## 三、Go实现验证

### 3.1 Formatter输出验证

Go的YunExpressFormatter必须输出符合Ruby缓存语义的结构：

```go
// domain/tracking/service/formatter.go
func (f *YunExpressFormatter) Format() *TrackingDetailDTO {
	// 1. 构建response包装层（关键）
	return &TrackingDetailDTO{
		Response: &TrackingResponseDTO{
			Item: TrackingItemDTO{
				TrackingNumber: trackingNumber,
				WayBillNumber:  wayBillNumber,
				TrackingStatus: latestStatus,
				PackageState:   packageState,
				// ... 其他字段
			},
		},
		SyncedAt: time.Now().UTC().Format(time.RFC3339), // ISO8601格式
	}
}
```

### 3.2 单元测试验证

单元测试必须验证：
1. `response.Item` 结构完整
2. `synced_at` 为ISO8601格式
3. JSON输出包含 `response` 和 `synced_at`

```go
// formatter_test.go
func TestFormatterOutputFormat(t *testing.T) {
	// 验证response包装层
	if dto.Response == nil {
		t.Fatal("Response wrapper is missing")
	}
	
	// 验证response.Item存在
	if dto.Response.Item.TrackingNumber == "" {
		t.Fatal("TrackingNumber missing in response.Item")
	}
	
	// 验证synced_at格式（ISO8601）
	_, err := time.Parse(time.RFC3339, dto.SyncedAt)
	if err != nil {
		t.Fatal("SyncedAt is not ISO8601 format")
	}
}
```

---

## 四、Ruby缓存不命中场景

**场景1：detail结构不完整**
```json
{
  "Item": { ... }  // 缺少response包装层 → Ruby不命中缓存
}
```

**场景2：synced_at过期**
```json
{
  "response": { "Item": { ... } },
  "synced_at": "2026-06-03T10:00:00Z"  // 超过4小时 → Ruby不命中缓存
}
```

**场景3：synced_at缺失**
```json
{
  "response": { "Item": { ... } }
  // 缺少synced_at → Ruby不命中缓存
}
```

---

## 五、验收标准

| 验收项 | 标准 | 验证方式 |
|--------|------|---------|
| **response.Item存在** | 必须包含TrackingNumber、WayBillNumber、TrackingStatus | 单元测试验证 |
| **synced_at格式** | ISO8601格式（RFC3339） | time.Parse验证 |
| **JSON输出结构** | 包含response + synced_at | JSON解析验证 |
| **Ruby缓存命中** | synced_at在4小时内 | 真实推送验证 |

---

## 六、真实推送验证流程

**Step 1：Go接收Webhook推送**
- 云途Webhook推送真实运单数据
- Go写入tracking_details（detail JSONB）

**Step 2：Ruby读取缓存**
- Ruby查询tracking_details
- Ruby执行fetch_cache方法
- 观察Ruby日志："Read from cache: YT..."

**Step 3：验证缓存命中**
- 如果命中 → Ruby返回缓存数据（不走Pull）
- 如果不命中 → Ruby调用Pull兜底（走云途API）

---

## 七、失败排查

**问题1：Ruby不命中缓存**
- 检查detail['response']['Item']是否完整
- 检查synced_at是否在4小时内
- 检查synced_at格式是否为ISO8601

**问题2：Ruby解析synced_at失败**
- Go写入的synced_at格式错误（不是ISO8601）
- synced_at为空字符串

**问题3：tracking_details写入失败**
- Go Upsert失败（数据库错误）
- detail字段为空

---

**文档版本**：v1.0
**创建时间**：2026-06-04
**对标Ruby**：tracking_detail.rb:246-252 (fetch_cache方法)