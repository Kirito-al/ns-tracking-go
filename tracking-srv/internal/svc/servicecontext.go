package svc

import (
	"time"

	"tracking-srv/internal/config"
	"tracking-srv/internal/dao"
	"tracking-srv/internal/event/dispatcher"
	queueconfig "tracking-srv/internal/queue/config"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ServiceContext 服务上下文（依赖注入容器）
type ServiceContext struct {
	Config config.Config

	// Redis 连接（缓存 + 任务队列）
	Redis *redis.Redis

	// GORM 数据库连接
	DB *gorm.DB

	// DAO 层（数据访问层）
	TrackingDAO         *dao.TrackingDAO         // tracking_details 表 DAO
	TrackingLogDAO      *dao.TrackingLogDAO      // tracking_logs 表 DAO
	OverseasPackageDAO  *dao.OverseasPackageDAO  // overseas_package_tracking_logs 表 DAO

	// 🆕 事件分发器（Demo 阶段新增）
	EventDispatcher dispatcher.Dispatcher

	// 🆕 Asynq Client（用于 Enqueue 任务）
	AsynqClient *asynq.Client
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 1. 创建 GORM 数据库连接
	db, err := gorm.Open(postgres.Open(c.DataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logx.Errorf("Failed to open database connection: %v", err)
		panic(err)
	}

	// 2. 配置连接池（通过 underlying sql.DB）
	sqlDB, err := db.DB()
	if err != nil {
		logx.Errorf("Failed to get underlying sql.DB: %v", err)
		panic(err)
	}

	if c.DatabaseConfig.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(c.DatabaseConfig.MaxOpenConns)
	}
	if c.DatabaseConfig.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(c.DatabaseConfig.MaxIdleConns)
	}
	if c.DatabaseConfig.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(c.DatabaseConfig.ConnMaxLifetime) * time.Second)
	}
	if c.DatabaseConfig.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(c.DatabaseConfig.ConnMaxIdleTime) * time.Second)
	}

	logx.Infof("PostgreSQL connected (GORM): %s", c.DataSource)

	// 3. 创建 Redis 连接
	var redisClient *redis.Redis
	if c.RedisConf.Host != "" {
		redisClient = redis.MustNewRedis(c.RedisConf)
		logx.Infof("Redis connected: %s", c.RedisConf.Host)
	} else {
		logx.Info("Redis not configured, cache disabled")
	}

	// 4. 创建 DAO 层
	trackingDAO := dao.NewTrackingDAO(db)
	trackingLogDAO := dao.NewTrackingLogDAO(db)
	overseasPackageDAO := dao.NewOverseasPackageDAO(db)

	// 5. 创建 Asynq Client（用于 Enqueue 任务）
	var asynqClient *asynq.Client
	if c.AsynqRedisConf.Addr != "" {
		asynqConfig := queueconfig.AsynqConfig{
			RedisAddr:     c.AsynqRedisConf.Addr,
			RedisPassword: c.AsynqRedisConf.Password,
			RedisDB:       c.AsynqRedisConf.DB,
		}
		asynqClient = queueconfig.NewAsynqClient(asynqConfig)
		logx.Infof("Asynq Client connected: addr=%s, db=%d", c.AsynqRedisConf.Addr, c.AsynqRedisConf.DB)
	} else {
		logx.Info("Asynq not configured, task queue disabled")
	}

	// 6. 创建事件分发器（默认为 nil，启动时注入）
	// 注意：EventDispatcher 需要在启动时注入（cmd/worker/main.go 或 tracking.go）
	// 这里先设置为 nil，后续通过 svcCtx.EventDispatcher = dispatcher.NewInMemoryDispatcher() 注入

	return &ServiceContext{
		Config:            c,
		Redis:             redisClient,
		DB:                db,
		TrackingDAO:       trackingDAO,
		TrackingLogDAO:    trackingLogDAO,
		OverseasPackageDAO: overseasPackageDAO,
		EventDispatcher:   nil, // 启动时注入
		AsynqClient:        asynqClient,
	}
}
