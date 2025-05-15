// @auth: pp
// @created: 2025-04-21 0:17
// @description: the operation of redis in user service
package user_utils

import (
	"context"
	"log"
	"strconv"
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

// @brief: 写入[uuid, token]
// @param: key: uuid
// @param: value: token
// @return: error
func (r *UserRedisSct) Set(ctx context.Context, key string, value any) error {
	err := r.RedisClient.Client.Set(ctx, key, value, 0).Err()
	if err != nil {
		log.Printf("Error setting key %s: %v", key, err)
		return err
	}
	return nil
}

// @brief: 写入[uuid, token]，并设置过期时间
// @param: key: uuid
// @param: value: token
// @param: expiration: 过期时间
// @return: error
func (r *UserRedisSct) SetWithTTL(ctx context.Context, key string, value any, expiration time.Duration) error {
	err := r.RedisClient.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {	
		log.Printf("Error setting key %s with expiration: %v", key, err)
		return err
	}
	return nil
}

func (r *UserRedisSct) Get(ctx context.Context, key string) (map[string]string, error) {
	val, err := r.RedisClient.Client.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Error getting key %s: %v", key, err)
		return nil, err
	}
	return map[string]string{key:val}, nil
}

func (r *UserRedisSct) Delete(ctx context.Context, key string) error {
	err := r.RedisClient.Client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Error deleting key %s: %v", key, err)
		return err
	}
	return nil
}

// @brief: 获取key:value以及过期时间
// @param: key: uuid
// @return: map[string]string: key:value
func (r *UserRedisSct) GetWithTTL(ctx context.Context, key string) (map[string]string, error){
	// Lua 脚本：原子性获取值和 TTL
	script := `
		local value = redis.call("GET", KEYS[1])
		local ttl = redis.call("TTL", KEYS[1])
		return {value, ttl}`
	// 执行 Lua 脚本
	rtl, err := r.RedisClient.Client.Do(ctx, "EVAL", script, 1, key).Result()
	if err != nil {
		log.Printf("Error executing Lua script for key %s: %v", key, err)
		return nil, err
	}
	log.Printf("the return result: %v", rtl)

	// 解析返回值
	rtlSlice, ok := rtl.([]any)
	if !ok || len(rtlSlice) != 2 {
		log.Printf("Invalid response from Lua script for (key=%s): %T", key, rtl)
		return nil, err
	}
	var valueStr string
	if rtlSlice[0] != nil {
		if str, ok := rtlSlice[0].(string); ok {
			valueStr = str
		} else {
			log.Printf("Invalid type for value: %T", rtlSlice[0])
			return nil, err
		}
	}
	// 解析 TTL
	ttl, ok := rtlSlice[1].(int64)
	if !ok {
		log.Printf("Invalid type for TTL: %T", rtlSlice[1])
		return nil, err
	}
	
	return map[string]string{
		key: valueStr,
		"ttl": strconv.FormatInt(ttl, 10),
	}, nil
}
