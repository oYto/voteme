package utils

import (
	"VoteMe/config"
	"VoteMe/control"
	"VoteMe/db"
	"VoteMe/model"
	"context"
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

// Init 初始化票据生成器
func init() {
	rand.Seed(time.Now().UnixNano())
	// 生成票据
	go ticketGenerator()
	getDbVotesToRedis()
	// 将redis中的数据累加到mysql中
	go syncVotesToDB()
}

// ticketGenerator是一个票据生成器，每20秒生成一个新的随机票据
// todo 这里为了方便测试，设置了20秒，后续改为需求中的2s
func ticketGenerator() {
	currentTicket = generateRandomString(10)
	ticker := time.NewTicker(config.TicketsUpdateTime)

	// 将当前有效票据写入 redis
	err := control.SetValidateTicket(currentTicket, config.MaxVotes, config.TicketsUpdateTime)
	if err != nil {
		log.Fatalf("createTicket to redis failed %s", err)
	}
	// 将当前有效的票据写入 mysql
	err = control.CreateOrTicket(currentTicket)
	if err != nil {
		log.Fatalf("createTicket to mysql failed %s", err)
	}

	for range ticker.C { // 循环监听定时器的通道
		ticketMutex.Lock()                       // 在修改 currentTicket 之前加锁
		currentTicket = generateRandomString(10) // 生成一个长度为10的随机字符串作为新票据

		// 将当前有效票据写入 redis
		err = control.SetValidateTicket(currentTicket, config.MaxVotes, config.TicketsUpdateTime)
		if err != nil {
			log.Fatalf("createTicket to redis failed %s", err)
		}
		// 将当前有效的票据写入 mysql
		err = control.CreateOrTicket(currentTicket)
		if err != nil {
			log.Fatalf("createTicket to mysql failed %s", err)
		}

		ticketMutex.Unlock() // 修改完成后解锁
	}
}

// generateRandomString函数生成一个指定长度的随机字符串
// 该字符串由小写字母、大写字母和数字组成
// todo 后续考虑使用更复杂的方式生成
func generateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n) // 创建一个长度为n的rune切片，用于存储生成的随机字符
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))] // 从letters中随机选择一个字符
	}
	return string(s) // 将rune切片转换为字符串并返回
}

// GetCurrentTicket GetCurrentTicket函数返回当前有效的票据
func GetCurrentTicket() string {
	return currentTicket // 返回当前有效的票据
}

func syncVotesToDB() {
	ticker := time.NewTicker(20 * time.Second) // 每分钟同步一次
	defer ticker.Stop()

	for range ticker.C {
		syncVotes()
	}
}

// 将redis中的票数同步到数据库中
func syncVotes() {
	// 获取所有需要同步的用户名列表
	userNames, err := getAllUserNames()
	if err != nil {
		log.Fatalf("getAllUserNames failed")
	}
	for _, userName := range userNames {
		key := fmt.Sprintf("votes:%s", userName)
		votes, err := db.GetRedisCLi().Get(context.Background(), key).Int()
		if err == redis.Nil {
			continue
		} else if err != nil {
			// 处理错误
			fmt.Println("Error getting votes from Redis:", err)
			continue
		}

		// 在这里更新数据库中的票数
		err = db.GetDB().Exec("UPDATE users SET votes =  ? WHERE name = ?", votes, userName).Error
		if err != nil {
			// 处理错误
			fmt.Println("Error updating votes in DB:", err)
			continue
		}

		// 同步成功后，重置Redis中的计数器
		//err = db.GetRedisCLi().DecrBy(context.Background(), key, int64(votes)).Err()
		//if err != nil {
		//
		//}
	}
}

// GetDbVotesToRedis 项目启动时，自动将数据库中的用户投票数据同步到Redis
func getDbVotesToRedis() error {
	var users []model.User

	// 从数据库中查询所有用户的name和votes字段
	if err := db.GetDB().Select("name", "votes").Find(&users).Error; err != nil {
		return err
	}

	ctx := context.Background()

	// 遍历用户数据，将每个用户的投票数同步到Redis
	for _, user := range users {
		// 使用用户的votes:name作为键，votes作为值
		key := fmt.Sprintf("votes:%s", user.Name)
		if err := db.GetRedisCLi().Set(ctx, key, user.Votes, 0).Err(); err != nil {
			return fmt.Errorf("failed to set Redis key for user %s: %v", user.Name, err)
		}
	}

	return nil
}

// 获取数据库中所有名字
func getAllUserNames() ([]string, error) {
	var users []model.User
	var userNames []string

	if err := db.GetDB().Find(&users).Error; err != nil {
		return nil, err
	}
	// 提取用户名
	for _, user := range users {
		userNames = append(userNames, user.Name)
	}

	return userNames, nil // 返回用户名列表
}
