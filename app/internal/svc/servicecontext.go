package svc

import (
	"ns-tracking-go/app/internal/config"
	"ns-tracking-go/domain/tracking/service"
	"ns-tracking-go/infrastructure/cache"
	"ns-tracking-go/infrastructure/database"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

// ServiceContext 服务上下文（DDD架构）
type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis
	DB     *gorm.DB

	// 领域服务
	TrackingService    *service.YunExpressFormatter
	TrackingLogService *service.TrackingLogService

	// 仓储实现
	TrackingRepo     *database.TrackingRepoImpl
	TrackingLogRepo  *database.TrackingLogRepoImpl
	OverseasPackageRepo *database.OverseasPackageRepoImpl

	// 缓存
	TrackingCache *cache.TrackingCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds := redis.MustNewRedis(redis.RedisConf{
		Host: c.RedisConf.Host,
		Pass: c.RedisConf.Pass,
		Type: "node",
	})

	db, err := database.NewDB(c.DataSource)
	if err != nil {
		panic(err)
	}

	// 复用 OverseasPackageRepoImpl 实例（避免重复创建）
	overseasPkgRepo := database.NewOverseasPackageRepoImpl(db)

	return &ServiceContext{
		Config: c,
		Redis:  rds,
		DB:     db.DB,
		TrackingService: service.NewYunExpressFormatter(nil),
		TrackingLogService: service.NewTrackingLogService(overseasPkgRepo),
		TrackingRepo: database.NewTrackingRepoImpl(db),
		TrackingLogRepo: database.NewTrackingLogRepoImpl(db),
		OverseasPackageRepo: overseasPkgRepo,
		TrackingCache: cache.NewTrackingCache(rds),
	}
}

func (ctx *ServiceContext) Close() {
	if ctx.DB != nil {
		sqlDB, _ := ctx.DB.DB()
		sqlDB.Close()
	}
}