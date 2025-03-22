// main.go
package main

import (
	"log"
	"star-go/api/routes"
	"star-go/pkg/core"
)

func main() {
	// 加载配置文件
	if err := core.InitConfig("./config.yaml"); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化日志系统
	if err := core.InitLogger(); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	// 初始化数据库连接
	if err := core.InitDatabase(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	// 初始化缓存 目前不用先预留
	// if err := core.InitCache(); err != nil {
	// 	log.Fatalf("初始化缓存失败: %v", err)
	// }

	// 初始化Gin引擎
	router := core.InitGin()

	// 设置路由
	routes.SetupRoutes(router)

	// 创建并运行应用程序
	app := core.NewApplication(router)
	app.Run()
}
