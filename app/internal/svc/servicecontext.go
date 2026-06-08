package svc

import (
	"ns-tracking-go/app/internal/config"
	"ns-tracking-go/domain/tracking/repo"
	"ns-tracking-go/domain/tracking/service"
	"ns-tracking-go/infrastructure/cache"
	"ns-tracking-go/infrastructure/database"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// ServiceContext 服务上下文（DDD架构）
type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis
	DB     *database.DB // ← 修复：改为 *database.DB（包装*gorm.DB）

	// 领域服务
	TrackingService    *service.YunExpressFormatter
	TrackingLogService *service.TrackingLogService

	// 仓储实现
	TrackingRepo     *database.TrackingRepoImpl
	TrackingLogRepo  *database.TrackingLogRepoImpl
	OverseasPackageRepo *database.OverseasPackageRepoImpl
	WebhookRequestLogRepo repo.WebhookRequestLogRepo // ← 修复：改为接口类型

	// 缓存
	TrackingCache *cache.TrackingCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	println("🔌 Connecting to Redis:", c.RedisHost)
	
	rds := redis.MustNewRedis(redis.RedisConf{
		Host: c.RedisHost,
		Pass: c.RedisPass,
		Type: c.RedisType,
	})
	
	println("✅ Redis connected")

	db, err := database.NewDB(c.DataSource)
	if err != nil {
		println("❌ DB connection failed:", err.Error())
		panic(err)
	}
	println("✅ DB connected")

	// 复用 OverseasPackageRepoImpl 实例（避免重复创建）
	overseasPkgRepo := database.NewOverseasPackageRepoImpl(db)
	webhookLogRepo := database.NewWebhookRequestLogRepoImpl(db)
	
	// 初始化TrackingCache
	trackingCache := cache.NewTrackingCache(rds)

	return &ServiceContext{
		Config: c,
		Redis:  rds,
		DB:     db,
		TrackingService: service.NewYunExpressFormatter(nil),
		TrackingLogService: service.NewTrackingLogService(overseasPkgRepo),
		TrackingRepo: database.NewTrackingRepoImpl(db, trackingCache), // ← 修复：注入TrackingCache
		TrackingLogRepo: database.NewTrackingLogRepoImpl(db),
		OverseasPackageRepo: overseasPkgRepo,
		WebhookRequestLogRepo: webhookLogRepo,
		TrackingCache: trackingCache,
	}
}

func (ctx *ServiceContext) Close() {
	if ctx.DB != nil {
		sqlDB, _ := ctx.DB.DB.DB() // ← 修复：三层调用（database.DB → *gorm.DB → sql.DB）
		sqlDB.Close()
	}
}