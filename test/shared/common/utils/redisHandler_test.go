package utils_test

import (
	"log"
	"testing"
	"context"
	"time"

	// "github.com/redis/go-redis/v9"
	user_config "walk/apps/user/config"
	utils "walk/shared/common/utils"
)

var redisClient *utils.RedisClient

func init() {
	// Initialize Redis configuration
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/user/config/user-service.yml"
	cfg, err := user_config.InitUserConfig(filePath)
	if err != nil {
		log.Fatalf("Failed to initialize user configuration: %v", err)
	}

	// Create a new Redis client
	redisClient, err = utils.NewRedisClient(utils.RedisConfig{
		Network:  cfg.Redis.Network,
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}

	log.Printf("Redis client created successfully: %v", redisClient)
}

func TestRedisClient(t *testing.T) {

	// 测试 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := redisClient.Client.Ping(ctx).Result(); err != nil {
		t.Errorf("failed to connect to Redis: %v", err)
	}
}