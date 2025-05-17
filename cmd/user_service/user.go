// Auth: pp
// Created: 
// Desc: 用户服务启动入口

package main

import (
	"fmt"
	"log"

	user_config "walk/apps/user/config"
	user_model "walk/apps/user/model"
	user_repo "walk/apps/user/repo"
	user_pb "walk/apps/user/rpc"
	user_service "walk/apps/user/service"
	user_utils "walk/apps/user/utils"

	utils "walk/shared/common/utils"

)

// @brief: 初始化配置
func init() {
	// 加载user_service.yaml

}

// @brief: 注册etcd: user
func registerEtcd() {
}

// @brief: 启动服务
func main() {

