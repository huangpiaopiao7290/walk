// 网关
// @auth: pp
// @date: 2025/05/15
// @desc: 网关服务

package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	etcd "go.etcd.io/etcd/clientv3"

	gw_config "walk/apps/gateway/config"
	handler "walk/apps/gateway/handler"
	interceptor "walk/apps/gateway/interceptor"
	// gw_pb "walk/apps/gateway/pb"
)

var (
	cfg *gw_config.GatewayConfig
	etcdClient *etcd.Client
)

// @brief: 初始化配置
func init() {
	// 读取配置
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/gw/config/grpc_gw_conf.yml"
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist at path: %s", filePath)
	}
	var err error
	cfg, err = gw_config.InitGatewayConfig(filePath)
	if err != nil {
		log.Fatalf("InitgwConfig returned unexpected error: %v", err)
	}
	// 初始化etcd
	etcdClient, err = etcd.New(etcd.Config{
		Endpoints:  	cfg.Etcd.Endpoints,
		DialTimeout: 	time.Duration(cfg.Etcd.DialTimeout) * time.Second,
	})
	if err != nil {
		panic(fmt.Sprintf("etcd.New failed: %v", err))
	}
}


func main() {
	// 创建DynamicHandler实例
	dynamicHandler := handler.NewDynamicHandler(etcdClient)

	// 创建Gin引擎
	r := gin.New()

	// 注册中间件
	r.Use(gin.Recovery())
	r.Use(interceptor.Logger())
	r.Use(interceptor.AuthRequired())

	// 注册所有endpoints
	for _, servver := range cfg.Servers {
		for _, endpoint := range servver.Endpoints {
			method := endpoint.Method
			path := endpoint.Path
			handlerFunc := dynamicHandler.HandleGRPCRequest(&endpoint)

			switch method {
			case http.MethodGet:
				r.GET(path, handlerFunc)
			case http.MethodPost:
				r.POST(path, handlerFunc)
			case http.MethodPut:
				r.PUT(path, handlerFunc)
			case http.MethodDelete:
				r.DELETE(path, handlerFunc)
			case http.MethodPatch:
				r.PATCH(path, handlerFunc)
			default:
				log.Printf("Unsupported method: %s", method)

			}
		}
	}

	// 启动HTTP服务
	log.Printf("Starting HTTP server on %s", cfg.Servers[0].Address)
	if err := r.Run(cfg.Servers[0].Address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}