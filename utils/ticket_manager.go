package utils

import (
	"VoteMe/config"
	"VoteMe/control"
	"VoteMe/db"
	"VoteMe/model"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"math/rand"
	"sync"
	"time"
)

var (
	currentTicket string     // 存储当前有效的票据
	ticketMutex   sync.Mutex // ticketMutex是一个互斥锁，用于控制对currentTicket变量的并发访问
)

func cleanTicketTable() {
	err := db.GetDB().Exec("TRUNCATE TABLE tickets").Error
	if err != nil {
		log.Fatalf("Failed to truncate table: %v", err)
	}
}

// DeleteKeysByPattern 删除redis中前缀符合pattern的键值对
func deleteKeysByPattern(pattern string) error {
	var cursor uint64
	var err error
	for {
		var keys []string
		keys, cursor, err = db.GetRedisCLi().Scan(context.Background(), cursor, pattern, 0).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			_, err = db.GetRedisCLi().Del(context.Background(), keys...).Result()
			if err != nil {
				return err
			}
		}
		if cursor == 0 { // 如果cursor为0，表示迭代完成
			break
		}
	}
	return nil
}

// ticketGenerator是一个票据生成器，每20秒生成一个新的随机票据
// todo 这里为了方便测试，设置了20秒，后续改为需求中的2s
func ticketGenerator() {
	var err error
	currentTicket, err = generateRandomHash(config.TicketLen)
	if err != nil {
		log.Fatalf("GenerateRandomHash failed %s", err)

	}
	ticker := time.NewTicker(config.TicketsUpdateTime)

	// 将当前有效票据写入 redis
	err = control.SetValidateTicket(currentTicket, config.MaxVotes, config.TicketsUpdateTime)
	if err != nil {
		log.Fatalf("createTicket to redis failed %s", err)
	}
	//将当前有效的票据写入 mysql
	err = control.CreateOrTicket(currentTicket)
	if err != nil {
		log.Fatalf("createTicket to mysql failed %s", err)
	}
	// 过期后，在这里重新生成票据
	for range ticker.C { // 循环监听定时器的通道
		ticketMutex.Lock()                                        // 在修改 currentTicket 之前加锁
		currentTicket, err = generateRandomHash(config.TicketLen) // 生成一个长度为10的随机字符串作为新票据
		if err != nil {
			log.Fatalf("GenerateRandomHash failed：%s", err)
		}
		// 将当前有效票据写入 redis
		err = control.SetValidateTicket(currentTicket, config.MaxVotes, config.TicketsUpdateTime)
		if err != nil {
			log.Fatalf("createTicket to redis failed %s", err)
		}
		//// 将当前有效的票据写入 mysql
		err = control.CreateOrTicket(currentTicket)
		if err != nil {
			log.Fatalf("createTicket to mysql failed %s", err)
		}

		ticketMutex.Unlock() // 修改完成后解锁
	}
}

func generateRandomHash(n int) (string, error) {
	// 生成足够的随机字节，由于转换成16进制后长度会翻倍，所以这里除以2
	bytes := make([]byte, n/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err // 在随机字节生成过程中返回错误
	}
	// 将字节序列转换为16进制字符串
	hash := hex.EncodeToString(bytes)
	// 如果需要的长度是奇数，则从生成的随机字符串中取前n个字符
	if len(hash) > n {
		hash = hash[:n]
	}
	return hash, nil
}

// generateRandomString函数生成一个指定长度的随机字符串
// 该字符串由小写字母、大写字母和数字组成
// todo 后续考虑使用更复杂的方式生成
//func generateRandomString(n int) string {
//	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
//
//	s := make([]rune, n) // 创建一个长度为n的rune切片，用于存储生成的随机字符
//	for i := range s {
//		s[i] = letters[rand.Intn(len(letters))] // 从letters中随机选择一个字符
//	}
//	return string(s) // 将rune切片转换为字符串并返回
//}

// GetCurrentTicket GetCurrentTicket函数返回当前有效的票据
func GetCurrentTicket() string {
	return currentTicket // 返回当前有效的票据
}

// 将redis中的票数同步到数据库中
func syncVotes() {
	// 获取所有需要同步的用户名列表
	userNames, err := getAllUserNames()
	if err != nil {
		log.Fatalf("getAllUserNames failed")
	}
	// 将 redis 中的 votes 逐个刷入mysql
	for _, userName := range userNames {
		key := fmt.Sprintf("Voteme:votes:%s", userName)
		votes, err := db.GetRedisCLi().Get(context.Background(), key).Int()
		if votes == 0 {
			continue
		}
		if err == redis.Nil {
			continue
		} else if err != nil {
			// 处理错误
			fmt.Println("Error getting votes from Redis:", err)
			continue
		}

		// 在这里更新数据库中的票数
		// 开启一个事务
		//tx := db.GetDB().Begin()
		//
		//// 尝试锁定特定用户的记录
		//var votesCount int
		//err = tx.Raw("SELECT votes FROM users WHERE name = ? FOR UPDATE", userName).Scan(&votesCount).Error
		//if err != nil {
		//	tx.Rollback() // 如果出现错误，回滚事务
		//	fmt.Println("Locking failed during disk brushing:", err)
		//}
		//
		//// 执行更新操作
		//err = tx.Exec("UPDATE users SET votes = votes + ? WHERE name = ?", votes, userName).Error
		//if err != nil {
		//	tx.Rollback() // 如果更新失败，回滚事务
		//	fmt.Println("Failed to update when brushing the disc：", err)
		//}
		//
		//// 提交事务
		//err = tx.Commit().Error
		//if err != nil {
		//	fmt.Println("Failed to submit transaction when brushing disk：", err)
		//
		//}

		err = db.GetDB().Exec("UPDATE users SET votes = votes +  ? WHERE name = ?", votes, userName).Error
		if err != nil {
			// 处理错误
			fmt.Println("Error updating votes in DB:", err)
			continue
		}

		// 同步成功后，重置Redis中的计数器
		err = db.GetRedisCLi().DecrBy(context.Background(), key, int64(votes)).Err()
		if err != nil {

		}
	}
}

// 获取数据库中所有名字
func getAllUserNames() ([]string, error) {
	var userNames []string

	if err := db.GetDB().Model(&model.User{}).Select("name").Find(&userNames).Error; err != nil {
		return nil, err
	}

	return userNames, nil // 返回用户名列表
}

//func getAllUserNames() ([]string, error) {
//	var users []model.User
//	var userNames []string
//
//	if err := db.GetDB().Find(&users).Error; err != nil {
//		return nil, err
//	}
//	// 提取用户名
//	for _, user := range users {
//		userNames = append(userNames, user.Name)
//	}
//
//	return userNames, nil // 返回用户名列表
//}
