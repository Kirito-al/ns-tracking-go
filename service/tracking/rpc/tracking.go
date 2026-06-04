package main

import (
	"flag"
	"fmt"

	"ns-tracking-go/service/tracking/rpc/internal/config"
	"ns-tracking-go/service/tracking/rpc/internal/event"
	"ns-tracking-go/service/tracking/rpc/internal/event/dispatcher"
	"ns-tracking-go/service/tracking/rpc/internal/server"
	"ns-tracking-go/service/tracking/rpc/internal/svc"
"ns-tracking-go/service/tracking/rpc/tracking"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/tracking.yaml", "the config file")

func main() {
	flag.Parse()

	// 1. 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 2. 创建服务上下文（依赖注入）
	ctx := svc.NewServiceContext(c)

	// 3. 创建并注册事件分发器
	eventDispatcher := dispatcher.NewInMemoryDispatcher()
	event.RegisterListeners(eventDispatcher)
	ctx.EventDispatcher = eventDispatcher

	// 4. 创建 gRPC 服务
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		tracking.RegisterTrackingServiceServer(grpcServer, server.NewTrackingServer(ctx))

		// 注册 gRPC 反射服务（用于调试）
		reflection.Register(grpcServer)
	})

	defer s.Stop()

	// 4. 启动服务
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
