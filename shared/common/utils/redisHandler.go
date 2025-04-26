// @auth: pp
// @created: 2025-04-21 0:17
// @description: the operation of redis

package utils

import (
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

type RedisConfig struct {
	Network      string // 网络类型(tcp/unix)
	Addr         string // redis地址
	Username     string // redis用户名
	Password     string // redis密码
	DB           int    // redis数据库
	PoolSize     int    // 最大连接数
	MinIdleConns int    // 最小空闲连接数
	MaxIdleConns int    // 最大空闲连接数
	MaxRetries   int    // 最大重试次数
}

func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	// 创建redis客户端
	rdb := redis.NewClient(&redis.Options{
		Network:      config.Network,
		Addr:         config.Addr,
		Username:     config.Username,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxIdleConns: config.MaxIdleConns,
		MaxRetries:   config.MaxRetries,
	})

	return &RedisClient{Client: rdb}, nil
}


