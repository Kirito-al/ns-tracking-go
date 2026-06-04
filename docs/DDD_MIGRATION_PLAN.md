# DDD Architecture Migration Plan

> 重构 demo1-gozero（ns-tracking-go）到DDD单体架构

---

## 一、现有架构分析（57个核心文件）

### 1.1 API层（29个文件）- HTTP REST服务

| 模块 | 文件数 | 功能 |
|------|-------|------|
| config | 2 | 配置加载+灰度发布 |
| formatter | 7 | **已迁移domain/tracking/service** |
| handler | 6 | Webhook/Tracking/Admin/Public路由 |
| logic | 8 | 业务编排（调用RPC） |
| middleware | 1 | CORS+限流 |
| normalize | 2 | node_code_mapper（已迁移）+ timezone_calculator |
| svc | 1 | ServiceContext（依赖注入） |
| types | 1 | HTTP请求响应结构体 |
| utils | 1 | signature（已迁移infrastructure/webhook） |

**关键入口**：tracking-api.go（HTTP服务启动）

---

### 1.2 RPC层（26个文件）- gRPC服务

| 模块 | 文件数 | 功能 |
|------|-------|------|
| cache | 2 | Redis缓存（4小时TTL）+ 系统缓存 |
| config | 1 | 数据库+Redis配置 |
| dao | 3 | **已迁移infrastructure/database** |
| event | 7 | 异步事件系统（dispatcher/listener/define） |
| logic | 4 | Upsert/GetTracking/MarkError/TokenResolver |
| model | 2 | **已迁移domain/tracking/entity** |
| queue | 3 | Asynq异步任务（retry_handler/retry_task） |
| server | 1 | gRPC服务注册 |
| svc | 1 | ServiceContext（GORM+Redis+DAO） |
| utils | 1 | mask（已迁移pkg/toolx） |

**关键入口**：tracking.go（gRPC服务启动）

---

### 1.3 Worker层（2个文件）- 异步任务处理

| 文件 | 功能 |
|------|------|
| main.go | Worker启动入口 |
| routes.go | 任务路由注册 |

---

## 二、DDD架构迁移方案

### 2.1 目标架构（用户指定）

```
ns-tracking-go/                # 重构后项目名（保持原名）
├── go.work                    # Go工作区（管理所有子模块）
├── app/                       # 【唯一启动入口】go-zero单体服务
│   ├── go.mod                 # 仅依赖domain+pkg
│   ├── etc/
│   │   └── app.yaml           # 服务配置（端口/DB/Redis）
│   ├── internal/
│   │   ├── handler/           # HTTP处理器（迁移API层handler）
│   │   ├── logic/             # 业务编排（迁移API层logic）
│   │   ├── svc/               # ServiceContext（合并API+RPC）
│   │   ├── middleware/        # CORS+限流
│   │   └── config/            # 配置结构体
│   ├── app.api                # 接口定义（迁移tracking.api）
│   └── main.go                # 【唯一启动入口】
│
├── domain/                    # DDD领域根目录
│   └── tracking/              # 核心域：轨迹追踪（独立模块）
│       ├── go.mod             # 专属依赖
│       ├── entity/            # 领域实体（已迁移✅）
│       ├── repo/              # 仓储接口（已迁移✅）
│       ├── service/           # 领域服务（已迁移formatter✅）
│       ├── event/             # 领域事件（迁移RPC event）
│       └── queue/             # 领域任务（迁移RPC queue）
│
├── infrastructure/            # 基础设施层
│   ├── database/              # 数据库实现（已迁移DAO✅）
│   ├── cache/                 # Redis缓存（迁移RPC cache）
│   └── webhook/               # Webhook验签（已迁移✅）
│
├── pkg/                       # 公共工具（已迁移✅）
│   ├── toolx/                 # 脱敏工具
│   ├── constant/              # 常量
│   └── errorx/                # 错误封装
│
└── README.md
```

---

### 2.2 关键设计决策

| 决策 | 说明 |
|------|------|
| **单服务启动** | 废弃API+RPC双服务，合并到app单入口 |
| **领域独立性** | domain/tracking独立go.mod，无外部依赖 |
| **基础设施下沉** | database/cache/webhook统一到infrastructure |
| **Go Workspaces** | go.work管理4模块（app/domain/infrastructure/pkg） |

---

## 三、迁移任务清单（P0优先级）

### 3.1 需迁移文件（57个）

**已迁移**（✅ 27个）：
- domain/tracking/service: 8个formatter文件
- domain/tracking/entity: 4个model文件
- infrastructure/database: 3个DAO文件
- infrastructure/webhook: 1个signature文件
- pkg/toolx: 1个mask文件

**待迁移**（❌ 30个）：
- app/internal/handler: 6个handler
- app/internal/logic: 8个logic
- app/internal/middleware: 1个middleware
- app/internal/svc: 1个ServiceContext（合并API+RPC）
- app/internal/config: 2个config
- app/internal/types: 1个types
- domain/tracking/event: 7个event文件
- domain/tracking/queue: 3个queue文件
- infrastructure/cache: 2个cache文件
- app/main.go: 1个启动入口
- app/app.api: 1个接口定义
- normalize/timezone_calculator.go: 1个（补充到domain/tracking）

---

## 四、迁移步骤

### Phase 1: 补充domain/tracking缺失功能
1. 迁移event系统（7文件）→ domain/tracking/event
2. 迁移queue系统（3文件）→ domain/tracking/queue
3. 补充timezone_calculator → domain/tracking/service

### Phase 2: 补充infrastructure层
1. 迁移cache（2文件）→ infrastructure/cache

### Phase 3: 创建app启动层
1. 创建app目录结构（go.mod + etc + internal）
2. 迁移handler（6文件）→ app/internal/handler
3. 迁移logic（8文件）→ app/internal/logic
4. 合并ServiceContext → app/internal/svc
5. 迁移middleware → app/internal/middleware
6. 迁移config/types → app/internal
7. 创建app.api + main.go（唯一启动入口）

### Phase 4: 清理旧架构
1. 删除service/tracking目录（废弃双服务）
2. 更新go.work（仅保留app/domain/infrastructure/pkg）
3. 验证编译 + 功能测试

---

## 五、风险与待办

| 风险 | 影响 | 解决方案 |
|------|------|---------|
| gRPC废弃 | 原有gRPC调用需改为直接调用domain | app层直接调用domain.service |
| Worker独立 | 异步任务需独立进程启动 | 保留Worker入口或集成到app |
| 配置合并 | API+RPC配置需合并 | app.yaml统一配置 |
| Proto废弃 | tracking.pb.go不再使用 | 直接使用domain.entity |

---

## 六、下一步执行

**当前进度**：27/57已迁移（47%）

**立即执行**：
1. ✅ 补充domain/tracking/event系统
2. ✅ 补充domain/tracking/queue系统
3. ✅ 补充infrastructure/cache
4. ✅ 创建app启动层（合并API+RPC）

---

**文档版本**：v1.0
**创建时间**：2026-06-04
**作者**：Sisyphus (OhMyOpenCode)