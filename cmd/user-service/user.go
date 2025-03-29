package main

// import (
// 	"log"
// 	"fmt"
// 	"github.com/gin-gonic/gin"
// 	"net/http"

// 	user_config "walk/configs/user"
// 	utils "walk/pkg/utils"
// 	user_restful "walk/api/user"

// )

// func init() {
// 	// 加载用户服务配置
// 	userConfig, err := user_config.InitUserConfig("../configs/user/user-config.yaml")
// 	if err != nil {
// 		log.Fatalf("failed to load user config: %v", err)
// 	}

// 	log.Printf("user config: %+v", userConfig)

// 	// 创建数据库连接池
// 	db := utils.NewDBConnection(&utils.DBConfig{
// 		Host:     userConfig.Database.Host,
// 		Port:     fmt.Sprintf("%d", userConfig.Database.Port),
// 		User:     userConfig.Database.Username,
// 		Passwd:   userConfig.Database.Password,
// 		DBName:   userConfig.Database.DBname,
// 	})
// 	utils.SetupDBConnectionPool(db, 10, 5, 60)

// 	// 显示连接池状态
// 	utils.ShowPoolStatus(db)

// }

// func main() {

// 	// 创建gin实例
// 	router := gin.Default()

// 	// 创建路由组
// 	userGroup := router.Group("/user")
// 	{	
// 		// 用户注册
// 		userGroup.POST("/register", user_restful.Register)
// 		// 用户登录
// 		// 用户注销
// 		// 用户信息查询
// 		// 用户信息修改

// 	}


// 	log.Println("user service has started...")
// }