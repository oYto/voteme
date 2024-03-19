package main

import (
	"VoteMe/db"      // 导入自定义的db包，用于数据库操作，注意替换为实际的导入路径
	"VoteMe/graphql" // 导入自定义的graphql包，其中定义了GraphQL的schema，注意替换为实际的导入路径
	"log"            // 导入log包，用于记录日志
	"net/http"       // 导入net/http包，用于HTTP服务器的功能
	"runtime"

	"github.com/graphql-go/handler" // 导入graphql-go/handler包，用于处理GraphQL请求
	_ "net/http/pprof"
)

func main() {
	// 初始化数据库
	db.InitDB()
	db.InitRedis() // 初始化Redis

	// 定义GraphQL服务可用的查询和变更操作
	schema, err := graphql.NewGraphQLSchema()
	if err != nil {
		// 记录错误日志并终止程序
		log.Fatalf("failed to create new schema, error: %v", err)
	}
	go func() {
		runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪，block
		runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪，mutex
		http.ListenAndServe(":6060", nil)
		log.Println("pprof is running on port 6060")
	}()

	// handler会解析请求，执行对应的GraphQL操作，并返回结果
	h := handler.New(&handler.Config{
		Schema: &schema, // 设置handler使用的GraphQL schema
		Pretty: true,    // 设置返回的JSON数据格式化，便于阅读
	})

	http.Handle("/graphql", h)

	// 输出日志，表示服务正在运行
	log.Println("Now server is running on port 9090")

	// 使用http.ListenAndServe函数在8080端口启动HTTP服务器
	log.Fatal(http.ListenAndServe(":9090", nil))

}
