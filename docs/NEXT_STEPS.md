# 下一步开发优先级梳理

> **梳理日期**: 2026-06-05
> **当前状态**: DDD架构重构完成，核心逻辑已迁移，编译验证通过
> **参考文档**: 
> - 【代码】ns-tracking@YunExpressService云途服务Go语言重写 + Webhook改造评估(1).md
> - 【梳理】ns-tracking 云途链路与 Webhook 改造说明.md

---

## 一、当前完成度分析

### ✅ 已完成部分（Phase 0-1）

| 功能模块 | 完成度 | 备注 |
|---------|-------|------|
| **DDD架构搭建** | 100% | domain/infrastructure/app/pkg四层完成 |
| **YunExpressFormatter** | 90% | 核心格式化逻辑完成，字段映射已补充 |
| **StatusMapper** | 100% | 12个状态码映射表完整 |
| **TrackingLogService** | 100% | 时间戳同步逻辑完成 |
| **Webhook Handler** | 80% | 签名验证已实现，TIS Push转换完成 |
| **Upsert入库** | 100% | TrackingRepo并发控制完成 |
| **编译验证** | 100% | 所有模块编译通过 |
| **代码审查问题修复** | 100% | P0/P1问题全部修复 |

---

## 二、关键缺失部分（按优先级）

### 🔴 P0 优先级（一票否决项）

| 缺失功能 | 影响范围 | 实施位置 | 对标Ruby | 状态 |
|---------|---------|---------|----------|------|
| **AES-CBC Payload解密** | Webhook加密推送无法处理 | `infrastructure/crypto/` | tms-talk `_decrypt_if_needed` | ✅ **已完成** |
| **真实tisPushData样本验证** | 标准化层设计基于推测 | 沙箱环境测试 | 云途文档 | ⬜ 待实施 |
| **完整节点代码映射表** | 状态码转换不准确 | `domain/tracking/service/node_code_mapper.go` | YunExpressTrackFormatter | ⬜ 待实施 |
| **Worker服务路径修复** | Asynq无法启动 | `app/cmd/worker/main.go` | 引用路径错误 | ⬜ 待实施 |

---

### 🟠 P1 优先级（核心功能）

| 缺失功能 | 影响范围 | 实施位置 | 对标Ruby |
|---------|---------|---------|----------|
| **GORM DB连接池配置** | 生产环境性能问题 | `infrastructure/database/db.go` | ActiveRecord连接池 |
| **Redis缓存层完整实现** | 幂等去重未完整 | `infrastructure/cache/` | Redis幂等key |
| **公开API查询逻辑** | admin/public接口无法使用 | `app/internal/logic/public/` | TrackingController |
| **服务真实启动验证** | 无法确认服务能否运行 | 配置+依赖注入 | Rails服务启动 |


---

### 🟡 P2 优先级（质量保障）

| 缺失功能 | 影响范围 | 实施位置 | 备注 |
|---------|---------|---------|------|
| **统一单元测试** | 质量保障缺失 | 各模块_test.go | 目标覆盖率>80% |
| **监控埋点** | 生产问题无法发现 | Prometheus指标 | Grafana告警 |
| **降级策略验证** | Pull兜底未实现 | Ruby侧改造 | YUN_EXPRESS_FALLBACK_PULL |
| **灰度发布配置** | 流量切换无控制 | `app/internal/config/gray_scale.go` | 10%→30%→50%→100% |

---

## 三、下一步开发路线图

### 📋 Phase 2: P0功能补充（第2-3周）

**目标**: 补齐一票否决项，确保Webhook接收可用

#### 任务清单

**2.1 AES-CBC Payload解密实现**
```
位置: infrastructure/crypto/aes_decrypt.go
参考: tms-talk/app/services/logistics/adapters/yuntu.py:_decrypt_if_needed
算法: AES-256-CBC + PKCS7Padding
密钥: YunExpressEncryptKey (环境变量)
流程:
  1. 检测encrypt字段
  2. Base64解码
  3. 提取IV (前16字节)
  4. AES-CBC解密
  5. PKCS7去填充
  6. 返回明文JSON
```

**2.2 真实tisPushData样本验证**
```
位置: docs/tis_push_data_samples.json (记录真实payload)
步骤:
  1. 联系云途获取沙箱账号
  2. 配置Webhook接收URL (cpolar映射)
  3. 发送测试推送
  4. 记录真实payload结构
  5. 对比文档推测字段
  6. 更标准化层逻辑
```

**2.3 完整节点代码映射表**
```
位置: domain/tracking/service/node_code_mapper.go
参考: ns-tracking/app/services/yun_express_track_formatter.rb
映射表: 39个节点代码 → TrackingStatus
关键节点:
  - XD/ZC → 10 (揽收)
  - FIRST_MILE_ARRIVE → 10 (揽收)
  - INTERNATIONAL_SHIP → 20 (国际运输)
  - CUSTOMS_CLEARANCE → 20 (海关清关)
  - DELIVERED → 50 (签收)
  - MD/TT → 50 (签收)
```

**2.4 Worker服务路径修复**
```
位置: app/cmd/worker/main.go
问题: 引用旧路径 app/rpc/internal
修复: 更新为 app/internal 路径
验证: Asynq Worker 启动测试
```

---

### 📋 Phase 3: P1功能补充（第3-5周）

**目标**: 补齐核心功能，确保服务可运行

#### 任务清单

**3.1 GORM DB连接池配置**
```go
// infrastructure/database/db.go
type DBConfig struct {
    MaxIdleConns    int    // 最大空闲连接: 10
    MaxOpenConns    int    // 最大打开连接: 100
    ConnMaxLifetime time.Duration // 连接最大生命周期: 1小时
    ConnMaxIdleTime time.Duration // 空闲连接超时: 10分钟
}

func NewDBWithPool(dataSource string, config DBConfig) (*DB, error) {
    db, err := gorm.Open(postgres.Open(dataSource), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // 设置连接池参数
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    
    return &DB{DB: db}, nil
}
```

**3.2 Redis缓存层完整实现**
```
位置: infrastructure/cache/tracking_cache.go
功能:
  1. 幂等key管理 (upsert:{tracking_number}:{synced_at})
  2. 缓存过期控制 (1小时过期)
  3. 缓存命中率统计
  4. Redis连接池配置
对标: Ruby Redis幂等key逻辑
```

**3.3 公开API查询逻辑实现**
```
位置: app/internal/logic/public/public_query_logic.go
      app/internal/logic/admin/admin_get_tracking_logic.go
功能:
  1. 调用 TrackingRepo.FindByTrackingNumber
  2. 解析 JSONB detail 字段
  3. 返回标准化响应
对标: TrackingController#index
```

**3.4 服务真实启动验证**
```
步骤:
  1. 检查配置文件完整性 (app.yaml)
  2. 检查数据库连接可用性
  3. 检查Redis连接可用性
  4. 启动HTTP服务 (go run app/main.go)
  5. 测试健康检查接口 (GET /health)
  6. 测试Webhook接收 (POST /webhook/tis-push)
```

**3.5 Webhook payload兼容性处理**
```
位置: app/internal/logic/webhook/tis_push_logic.go
处理三种信封格式:
  格式1: { "customerCode": "...", "timestamp": "...", "body": {...} }
  格式2: { ... } (直接body)
  格式3: { "encrypt": "base64密文" } (AES加密)
参考: tms-talk/app/services/logistics/normalize.py
```

---

### 📋 Phase 4: P2功能补充（第5-7周）

**目标**: 质量保障和监控，确保生产可用

#### 任务清单

**4.1 统一单元测试**
```
覆盖模块:
  - domain/tracking/service/*_test.go
  - infrastructure/database/*_test.go
  - infrastructure/crypto/*_test.go
  - app/internal/logic/*_test.go
目标覆盖率: >80%
测试类型:
  - 单元测试 (纯函数测试)
  - 集成测试 (跨模块流程)
  - E2E测试 (真实Webhook推送)
```

**4.2 监控埋点实现**
```
位置: infrastructure/metrics/prometheus.go
指标清单:
  - webhook_received_total (Counter)
  - webhook_success_rate (Gauge)
  - upsert_duration_seconds (Histogram)
  - cache_hit_rate (Gauge)
  - idempotent_key_hit_total (Counter)
告警规则:
  - P0: Webhook失败率 > 5%
  - P1: Upsert耗时 > 500ms
  - P2: 缓存命中率 < 90%
```

**4.3 Ruby降级策略验证**
```
位置: D:\gohome\shenhe1\ns-tracking\app\services\yun_express_service.rb
改造:
  1. YunExpressService#call 改为读缓存优先
  2. 保留 Pull 兜底开关 (YUN_EXPRESS_FALLBACK_PULL)
  3. 对比 Go缓存 vs Ruby Pull 数据一致性
验证:
  - 缓存命中率统计
  - Pull兜底触发频率
  - 数据一致性验证
```

**4.4 灰度发布配置**
```
位置: app/internal/config/gray_scale.go
配置:
  - GrayScaleEnabled: bool
  - GrayScalePercentage: int (10/30/50/100)
流程:
  1. Go服务上线，接收Webhook写库
  2. Ruby还在Pull模式 (灰度开关开启)
  3. 10%流量走Go缓存，90%走Pull
  4. 观察1周 → 30% → 50% → 100%
监控:
  - Webhook接收成功率
  - 缓存命中率
  - 异常告警频率
```

---

## 四、关键风险点与缓解措施

### 🔥 高风险点

| 风险项 | 等级 | 缓解措施 |
|-------|------|---------|
| **真实payload与文档不一致** | P0 | 沙箱验证优先，记录真实样本 |
| **AES解密密钥不正确** | P0 | 云途申请密钥，沙箱验证 |
| **节点代码映射缺失** | P0 | Ruby完整映射表迁移，补充测试 |
| **Worker无法启动** | P0 | 路径修复，启动验证 |
| **缓存数据结构不一致** | P1 | Go vs Ruby对比测试，JSON结构验证 |

---

### ⚠️ 中风险点

| 风险项 | 筺级 | 缓解措施 |
|-------|------|---------|
| **数据库连接池配置不当** | P1 | 参考 ActiveRecord 连接池经验值 |
| **Redis幂等key过期时间** | P1 | 1小时窗口，验证去重效果 |
| **公开API无法使用** | P1 | 补充Logic层实现，测试验证 |
| **服务启动失败** | P1 | 依赖注入检查，健康检查接口 |
| **降级策略失效** | P2 | Pull兜底开关保留，灰度验证 |

---

## 五、验收标准

### P0验收标准（Phase 2完成）

```
✅ AES解密成功率 > 99% (加密payload解析)
✅ 节点代码映射覆盖率 > 95% (关键节点全覆盖)
✅ Worker服务启动成功 (Asynq运行)
✅ 真实payload样本记录 (沙箱验证)
```

---

### P1验收标准（Phase 3完成）

```
✅ DB连接池配置生效 (生产负载测试)
✅ Redis幂等去重命中率 > 90%
✅ 公开API返回正确数据 (缓存命中)
✅ 服务启动健康检查通过 (GET /health)
✅ Webhook三种格式兼容解析
```

---

### P2验收标准（Phase 4完成）

```
✅ 单元测试覆盖率 > 80%
✅ Prometheus指标采集正常
✅ Grafana告警触发测试通过
✅ Ruby降级策略验证通过
✅ 灰度发布流程验证 (10%→100%)
```

---

## 六、总结

### 核心结论

1. **当前状态**: DDD架构重构完成，核心逻辑已迁移，但关键功能缺失
2. **下一步优先级**: P0功能补充（AES解密+真实payload验证+节点映射+Worker修复）
3. **预计完成时间**: 
   - Phase 2 (P0): 第2-3周
   - Phase 3 (P1): 第3-5周
   - Phase 4 (P2): 第5-7周
4. **验收标准**: 按Phase验收，确保每阶段功能可用

### 关键建议

1. **沙箱验证优先**: 真实payload样本是标准化层设计的基础
2. **AES解密关键**: Webhook加密推送无法跳过
3. **节点映射完整**: 状态码转换准确性影响缓存命中
4. **Worker服务修复**: Asynq是失败重试的关键机制
5. **分阶段验收**: 每Phase完成后验证，避免后期问题堆积

---

**文档版本**: v2.0
**最后更新**: 2026-06-05
**梳理人**: Sisyphus (OhMyOpenCode)