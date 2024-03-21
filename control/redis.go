package control

import (
	"VoteMe/config"
	"VoteMe/db"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var ctx = context.Background()

// UpdateUserVotesWithLock redis 分布式锁进行投票
func UpdateUserVotesWithLock(userName string) error {
	lockKey := "Voteme:update:user:vote:lock:" + userName
	lockVal := "1" // 用于标识锁的持有者，可以是一个更复杂的标识，如UUID

	startTime := time.Now()
	maxTime := 10 * time.Second
	// 循环直到获取锁或超过最大等待时间
	for {
		if time.Since(startTime) > maxTime {
			return fmt.Errorf("failed to acquire lock for user %s within %d seconds", userName, maxTime)
		}

		// 尝试获取锁
		locked, err := db.GetRedisCLi().SetNX(ctx, lockKey, lockVal, 10*time.Millisecond).Result()
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
				_, err := db.GetRedisCLi().Eval(ctx, script, []string{lockKey}, lockVal).Result()
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

// SetValidateTicket 将有效票据缓存起来，设置过期时间以及使用次数
func SetValidateTicket(ticketID string, maxVotes int, ticketUpdateTime time.Duration) error {
	//maxVotesStr := fmt.Sprint(maxVotes)
	ticketIDCache := fmt.Sprintf("Voteme:ticketIDCache:%s", ticketID)
	err := db.GetRedisCLi().Set(context.Background(), ticketIDCache, maxVotes, ticketUpdateTime).Err()
	if err != nil {
		return err
	}
	return nil
}

// DecreaseUsageLimit 减少键的使用次数，并检查是否达到上限或过期
func DecreaseUsageLimit(ticketID string) error {
	ticketIDCache := fmt.Sprintf("Voteme:ticketIDCache:%s", ticketID)

	// 使用DECR命令减少票据的可用次数
	result, err := db.GetRedisCLi().Decr(context.Background(), ticketIDCache).Result()
	if err != nil {
		return err // 处理可能的Redis错误
	}

	if result < 0 {
		// 票据使用次数已超上限
		return fmt.Errorf("ticket %s has reached its maximum usage", ticketID)
	}

	// 票据有效
	return nil
}

// GetVotesByName 获取某个选手的票数：这里是缓存，会有一定时延,导致数据不准确
func GetVotesByName(name string) (int, error) {
	key := fmt.Sprintf("Voteme:current:votes:%s", name)
	votesStr, err := db.GetRedisCLi().Get(context.Background(), key).Result()
	if err == redis.Nil {
		lockKey := "Voteme:get:user:vote:lock:" + name // 使用不同的键作为锁
		lockValue := "1"
		// 尝试获取锁
		ok, err := db.GetRedisCLi().SetNX(context.Background(), lockKey, lockValue, 20*time.Millisecond).Result()
		if err != nil {
			return 0, err
		}

		if ok {
			defer db.GetRedisCLi().Del(context.Background(), lockKey) // 确保释放锁
			votes, err := GetUserVotes(name)
			if err != nil {
				return 0, err
			}
			//fmt.Println("hit mysql---------")
			db.GetRedisCLi().Set(context.Background(), key, votes, config.TicketCacheRefreshTime)
			return votes, nil
		}

		// 如果没有获取到锁，则等待一段时间后重试
		for i := 0; i < 3; i++ { // 重试次数
			time.Sleep(10 * time.Millisecond)                                         // 等待时间
			votesStr, err = db.GetRedisCLi().Get(context.Background(), name).Result() // 尝试再次从缓存获取
			if err == nil {
				GetVotesByName(name) // 递归获取或者直接break
			}
		}
	}

	if err != nil && err != redis.Nil {
		return 0, err
	}

	votesInt, err := strconv.Atoi(votesStr)
	if err != nil {
		return 0, err
	}
	//fmt.Println("hit redis")
	return votesInt, nil
}

func VoteForUserRedis(userName string) error {
	// 投票计数器的键
	key := fmt.Sprintf("Voteme:votes:%s", userName)
	// 增加用户的票数
	_, err := db.GetRedisCLi().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

// SetTicketUsageLimitInRedis 在Redis中设置票据的使用上限和过期时间
//func SetTicketUsageLimitInRedis(ticketID string, limit int) error {
//	// 假设使用Redis客户端rdb和上下文ctx
//	_, err := db.GetRedisCLi().Set(context.Background(), ticketID, limit, config.TicketsUpdateTime).Result()
//	return err
//}

// GetVotesByName 可能存在缓存穿透
//func GetVotesByName(name string) (int, error) {
//	votesStr, err := db.GetRedisCLi().Get(context.Background(), name).Result()
//	if err == redis.Nil {
//		votes, err := GetUserVotes(name)
//		//fmt.Println("hit mysql")
//		if err != nil {
//			return 0, err
//		}
//		err = db.GetRedisCLi().Set(context.Background(), name, votes, config.TicketCacheRefreshTime).Err()
//		if err != nil {
//			return 0, err
//		}
//		return votes, nil
//	} else if err != nil {
//		return 0, err
//	}
//
//	votesInt, err := strconv.Atoi(votesStr)
//	if err != nil {
//		return 0, err
//	}
//	//fmt.Println("hit redis")
//	return votesInt, nil
//}
