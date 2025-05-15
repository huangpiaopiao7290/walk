package user_utils_test

import (
	"log"
	"os"
	"testing"
	"context"


	"walk/apps/user/config"
	"walk/apps/user/utils"
	"walk/shared/common/utils"
)

var (
	userConfig *user_config.UserConfig
	userRedisSct *user_utils.UserRedisSct
)

func init() {
	// 初始化配置信息
	filePath := "/home/pp/programs/program_go/timeTrack/walk/apps/user/config/user_service.yml"
	// Ensure the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("The file %s does not exist", filePath)
	}
	var err error
	userConfig, err = user_config.InitUserConfig(filePath)
	if err != nil {
		log.Printf("Failed to initialize user configuration: %v", err)
	}

	// 初始化redis配置
	rdsConfig := &utils.RedisConfig{
		Network:  userConfig.Redis.Network,
		Addr:     userConfig.Redis.Addr,
		Username: userConfig.Redis.Username,
		Password: userConfig.Redis.Password,
		DB:       userConfig.Redis.DB,
		PoolSize: userConfig.Redis.PoolSize,
		MinIdleConns: userConfig.Redis.MinIdleConns,
		MaxIdleConns: userConfig.Redis.MaxIdleConns,
		MaxRetries: userConfig.Redis.MaxRetries,
	}
	log.Printf("Mapped RedisConfig: %+v", rdsConfig)

	// 创建Redis客户端
	rds, err := utils.NewRedisClient(*rdsConfig)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	// 初始化userRedisSct
	userRedisSct = user_utils.NewRedisHandler(rds)

	log.Printf("UserRedisSct: %+v", userRedisSct)

	if userRedisSct == nil {
		log.Fatalf("Failed to create UserRedisSct: %v", err)
	}
}

func TestUserRedisSct_Set(t *testing.T) {
	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	err := userRedisSct.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}
}
