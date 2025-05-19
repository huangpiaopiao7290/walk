// Auth: pp
// Created:
// Desc: 用户服务启动入口

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	user_config "walk/apps/user/config"
	user_model "walk/apps/user/model"
	user_repo "walk/apps/user/repo"
	user_pb "walk/apps/user/rpc"
	user_service "walk/apps/user/service"
	user_utils "walk/apps/user/utils"

	"go.etcd.io/etcd/clientv3"
	"gorm.io/gorm"

	utils "walk/shared/common/utils"
)

var (
	db 			 *gorm.DB
	cfg 		 *user_config.UserConfig
	userRepo 	 user_repo.UserRepo[user_model.User]
	userRedisSct *user_utils.UserRedisSct
	etcdClient 	 *clientv3.Client
)

// @brief: 初始化配置
func init() {
	// 加载user_service.yaml
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/user/config/user_service.yml"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("The file %s does not exist", filePath)
	}

	var err error
	cfg, err = user_config.InitUserConfig(filePath)
	if err != nil {
		log.Printf("Failed to initialize user configuration: %v", err)
	}
	// 初始化DB
	db, err = utils.NewDBConnection(&utils.DBConfig{
		Host:   cfg.Database.Host,
		Port:   fmt.Sprintf("%d", cfg.Database.Port),
		User:   cfg.Database.Username,
		Passwd: cfg.Database.Password,
		DBName: cfg.Database.DBname,
	})
	if err != nil {
		log.Printf("Failed to create database connection: %v", err)
	}
	log.Printf("Database connection information: %v", db)

	// 配置连接池
	err = utils.SetupDBConnectionPool(db, 10, 5, 60)
	if err != nil {
		log.Fatalf("Failed to configure connection pool: %v", err)
	}

	// 初始化 userRepo
	userRepo = user_repo.NewUserRepo[user_model.User](db)

	// 初始化userRedisSct
	rdsConfig := &utils.RedisConfig{
		Network:  cfg.Redis.Network,
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxIdleConns: cfg.Redis.MaxIdleConns,
		MaxRetries: cfg.Redis.MaxRetries,
	}
	rds, err := utils.NewRedisClient(*rdsConfig)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	userRedisSct = user_utils.NewRedisHandler(rds)
	if userRedisSct == nil {
		log.Fatalf("Failed to create UserRedisSct: %v", err)
	}

	// 初始化etcd
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   cfg.Etcd.Endpoints,
		DialTimeout: time.Duration(cfg.Etcd.DialTimeout) * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
}

// @brief: 注册etcd for user
func registerUserSevice2Etcd(serviceName, addr string, ttl int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Etcd.RequestTimeout)*time.Second)
	defer cancel()

	// 创建租约
	leaseResp, err := etcdClient.Grant(ctx, ttl)
	if err != nil {
		return fmt.Errorf("failed to create lease: %v", err)
	}
	// 注册服务
	_, err = etcdClient.Put(ctx, serviceName, addr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}
	// 自动续约
	ch, err := etcdClient.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return fmt.Errorf("failed to keep alive: %v", err)
	}
	go func() {
		for {
			select {
			case resp := <-ch:
				if resp == nil {
					log.Printf("Lease %d expired", leaseResp.ID)
					return
				}
				log.Printf("Lease %d renewed", resp.ID)
			}
		}
	}()
	return nil

}

// @brief: 启动服务
func main() {
	port := cfg.Server.Port
	
}
