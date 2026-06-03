package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"tracking-srv/internal/config"
	"tracking-srv/internal/event"
	"tracking-srv/internal/event/dispatcher"
	queueconfig "tracking-srv/internal/queue/config"
	"tracking-srv/internal/queue/handler"
	"tracking-srv/internal/svc"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/tracking.yaml", "the config file")

func main() {
	flag.Parse()

	// 1. 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 2. 创建 ServiceContext（包含 DAO、Redis、Dispatcher）
	svcCtx := svc.NewServiceContext(c)

	// 3. 注册事件监听器
	eventDispatcher := dispatcher.NewInMemoryDispatcher()
	event.RegisterListeners(eventDispatcher)
	svcCtx.EventDispatcher = eventDispatcher

	// 4. 创建 Asynq Server（Worker 进程）
	asynqConfig := queueconfig.AsynqConfig{
		RedisAddr:     c.RedisConf.Host,
		RedisPassword: c.RedisConf.Pass,
		RedisDB:       c.AsynqRedisConf.DB,
		Concurrency:   c.AsynqRedisConf.Concurrency,
	}
	asynqServer := queueconfig.NewAsynqServer(asynqConfig)

	// 5. 注册任务处理器（路由）
	mux := asynq.NewServeMux()
	handler.RegisterRetryHandler(mux, svcCtx)

	// 下期预留：
	// handler.RegisterLogHandler(mux, svcCtx)
	// handler.RegisterKafkaHandler(mux, svcCtx)

	// 6. 启动 Worker 进程
	logx.Info("Asynq Worker starting...")
	go func() {
		if err := asynqServer.Run(mux); err != nil {
			logx.Errorf("Asynq Worker failed: %v", err)
			os.Exit(1)
		}
	}()

	logx.Info("Asynq Worker started successfully")

	// 7. 等待中断信号（优雅关闭）
	waitForShutdown(asynqServer)
}

// waitForShutdown 等待中断信号（优雅关闭）
func waitForShutdown(server *asynq.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logx.Infof("Received signal: %v, shutting down...", sig)

	// 关闭 Asynq Server
	server.Shutdown()
	logx.Info("Asynq Worker shutdown completed")
}