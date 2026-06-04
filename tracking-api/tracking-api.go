package main

import (
	"flag"
	"fmt"

	"tracking-api/internal/config"
	routes "tracking-api/internal/handler"
	"tracking-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/tracking-api-api.yaml", "the config file")

func main() {
	flag.Parse()

	// 1. 加载配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 2. 创建 HTTP 服务
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 3. 创建服务上下文（依赖注入）
	ctx := svc.NewServiceContext(c)
	defer ctx.Close() // 确保资源释放

	// 4. 注册 HTTP 处理器
	routes.RegisterHandlers(server, ctx)

	// 5. 启动服务
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}