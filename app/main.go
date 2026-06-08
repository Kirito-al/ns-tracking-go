package main

import (
	"flag"

	"ns-tracking-go/app/internal/config"
	routes "ns-tracking-go/app/internal/handler"
	"ns-tracking-go/app/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/app.yaml", "the config file")

func main() {
	flag.Parse()

	// 1. 加载配置
	c, err := config.Load(*configFile)
	if err != nil {
		panic(err)
	}

	println("✅ Config loaded:")
	println("  Host:", c.Host)
	println("  Port:", c.Port)
	println("  Redis:", c.RedisHost)

	// 2. 创建 HTTP 服务
	server := rest.MustNewServer(rest.RestConf{
		Host: c.Host,
		Port: c.Port,
	})
	defer server.Stop()

	println("✅ Server created on", c.Host+":"+string(rune(c.Port+'0')))

	// 3. 创建服务上下文（依赖注入）
	ctx := svc.NewServiceContext(*c)
	defer ctx.Close()

	println("✅ ServiceContext created")

	// 4. 注册路由
	routes.RegisterHandlers(server, ctx)

	println("✅ Handlers registered, starting server...")

	// 5. 启动服务
	server.Start()
}