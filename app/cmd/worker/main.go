package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"ns-tracking-go/app/internal/config"
	"ns-tracking-go/app/internal/svc"
	queueconfig "ns-tracking-go/domain/tracking/queue/config"
	"ns-tracking-go/app/internal/queue/handler"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
)

var configFile = flag.String("f", "etc/tracking.yaml", "the config file")

func main() {
	flag.Parse()

	// 1. 加载配置
	c, err := config.Load(*configFile)
	if err != nil {
		logx.Errorf("Load config failed: %v", err)
		os.Exit(1)
	}

	// 2. 创建 ServiceContext（包含 DAO、Redis）
	svcCtx := svc.NewServiceContext(*c)

	// 3. 创建 Asynq Server（Worker 进程）
	asynqConfig := queueconfig.AsynqConfig{
		RedisAddr:     c.AsynqRedisHost,    // ← 修复：使用AsynqRedisHost
		RedisPassword: c.AsynqRedisPass,    // ← 修复：使用AsynqRedisPass
		RedisDB:       c.AsynqRedisDB,      // ← 修复：使用AsynqRedisDB
		Concurrency:   c.AsynqConcurrency,  // ← 修复：使用AsynqConcurrency
	}
	asynqServer := queueconfig.NewAsynqServer(asynqConfig)

	// 4. 注册任务处理器（路由）
	mux := asynq.NewServeMux()
	handler.RegisterRetryHandler(mux, svcCtx)           // ← 保留：重试任务
	handler.RegisterTisPushAsyncHandler(mux, svcCtx)    // ← 新增：TIS Push异步处理

	// 下期预留：
	// handler.RegisterLogHandler(mux, svcCtx)
	// handler.RegisterKafkaHandler(mux, svcCtx)

	// 6. 启动 Worker 进程（用 errgroup + context 控制生命周期）
	logx.Info("Asynq Worker starting...")
	
	// 创建可取消的 context（关键：必须用 WithCancel 才能主动触发取消）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保 cancel 被调用（资源释放）
	
	g, ctx := errgroup.WithContext(ctx)
	
	// 启动 Asynq Server
	g.Go(func() error {
		logx.Info("Asynq Worker started successfully")
		return asynqServer.Run(mux)
	})
	
	// 监听 context 取消信号（优雅关闭）
	g.Go(func() error {
		<-ctx.Done()
		logx.Info("Context cancelled, shutting down Asynq Worker...")
		asynqServer.Shutdown()
		return nil
	})
	
	// 等待中断信号，触发优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		logx.Infof("Received signal: %v, cancelling context...", sig)
		
		// 关键：调用 cancel() 触发 errgroup 取消（而不是 ctx.Done()）
		cancel() // 触发 errgroup 取消 → 优雅关闭 Asynq Worker
	}()
	
	// 等待所有 goroutine 完成
	if err := g.Wait(); err != nil {
		logx.Errorf("Asynq Worker exited with error: %v", err)
		os.Exit(1)
	}
	logx.Info("Asynq Worker shutdown completed")
}