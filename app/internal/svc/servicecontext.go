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

	return &ServiceContext{
		Config: c,
		Redis:  rds,
		DB:     db.DB,
		TrackingService: service.NewYunExpressFormatter(nil),
		TrackingLogService: service.NewTrackingLogService(database.NewOverseasPackageRepoImpl(db)),
		TrackingRepo: database.NewTrackingRepoImpl(db),
		TrackingLogRepo: database.NewTrackingLogRepoImpl(db),
		OverseasPackageRepo: database.NewOverseasPackageRepoImpl(db),
		TrackingCache: cache.NewTrackingCache(rds),
	}
}

func (ctx *ServiceContext) Close() {
	if ctx.DB != nil {
		sqlDB, _ := ctx.DB.DB()
		sqlDB.Close()
	}
}