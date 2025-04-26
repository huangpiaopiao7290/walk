// @auth: pp
// @created: 2025-04-21 0:17
// @description: the operation of redis in user service
package user_utils

import (
	"context"
	"log"
	"time"

	"walk/shared/common/utils"
)

type UserRedisSct struct {
	RedisClient *utils.RedisClient
}

func NewRedisHandler(redisClient *utils.RedisClient) *UserRedisSct {
	return &UserRedisSct{
		RedisClient: redisClient,
	}
}

func (r *UserRedisSct) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	err := r.RedisClient.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Error setting key %s: %v", key, err)
		return err
	}
	return nil
}

func (r *UserRedisSct) Get(ctx context.Context, key string) (string, error) {
	val, err := r.RedisClient.Client.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Error getting key %s: %v", key, err)
		return "", err
	}
	return val, nil
}

func (r *UserRedisSct) Delete(ctx context.Context, key string) error {
	err := r.RedisClient.Client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Error deleting key %s: %v", key, err)
		return err
	}
	return nil
}
