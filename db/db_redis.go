package db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"math/rand"
	"time"
)

var Rdb *redis.Client
var ctx = context.Background()

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "47.92.151.211:16379", // Redis地址
		Password: "",                    // 密码
		DB:       0,                     // 默认数据库
	})

	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

func UpdateUserVotesWithLock(userName string) error {
	lockKey := "lock:user:" + userName
	lockVal := "1" // 用于标识锁的持有者，可以是一个更复杂的标识，如UUID

	startTime := time.Now()
	maxTime := 10 * time.Second
	// 循环直到获取锁或超过最大等待时间
	for {
		if time.Since(startTime) > maxTime {
			return fmt.Errorf("failed to acquire lock for user %s within %d seconds", userName, maxTime)
		}

		// 尝试获取锁
		locked, err := Rdb.SetNX(ctx, lockKey, lockVal, 20*time.Millisecond).Result()
		if err != nil {
			return fmt.Errorf("error while attempting to lock for user %s: %v", userName, err)
		}
		if locked {
			// 成功获取锁，使用defer语句确保最后释放锁
			defer func() {
				// 使用Lua脚本来安全释放锁
				script := `
                if redis.call("get", KEYS[1]) == ARGV[1] then
                    return redis.call("del", KEYS[1])
                else
                    return 0
                end`
				_, err := Rdb.Eval(ctx, script, []string{lockKey}, lockVal).Result()
				if err != nil {
					log.Printf("failed to release lock for user %s: %v\n", userName, err)
				}
			}()

			return UpdateUserVotes(userName) // 调用原有逻辑更新票数
		}

		// 使用一个更大的随机间隔来减少锁竞争
		time.Sleep(time.Duration(rand.Intn(100)+10) * time.Millisecond) // 随机等待时间在100到600毫秒之间

	}
}
