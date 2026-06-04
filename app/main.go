package main

import (
	"flag"

	"ns-tracking-go/app/internal/config"
	routes "ns-tracking-go/app/internal/handler"
	"ns-tracking-go/app/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/app.yaml", "the config file")

func main() {
	flag.Parse()

	// 1. 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 2. 创建 HTTP 服务（单体服务）
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 3. 创建服务上下文（依赖注入）
	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	// 4. 注册路由
	routes.RegisterHandlers(server, ctx)

	// 5. 启动服务
	server.Start()
}