package db

import (
	"VoteMe/config"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"sync"
)

var (
	redisConn *redis.Client
	redisOnce sync.Once
)

func initRedis() {
	redisConfig := config.GetGlobalConf().RedisConfig
	addr := fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)
	redisConn = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: redisConfig.PassWord,
		DB:       redisConfig.DB,
		PoolSize: redisConfig.PoolSile,
	})

	// 连接测试以确保与 Redis 服务器的通信正常。
	_, err := redisConn.Set(context.Background(), "abc", 100, 60).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}
func GetRedisCLi() *redis.Client {
	redisOnce.Do(initRedis)
	return redisConn
}
