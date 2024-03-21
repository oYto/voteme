package utils

import (
	"VoteMe/config"
	"VoteMe/control"
	"VoteMe/db"
	"context"
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
	go ticketGenerator()
}

// ticketGenerator是一个票据生成器，每20秒生成一个新的随机票据
// todo 这里为了方便测试，设置了20秒，后续改为需求中的2s
func ticketGenerator() {
	currentTicket = generateRandomString(10)
	ticker := time.NewTicker(config.TicketsUpdateTime)

	//// 将当前有效票据写入 redis
	//err := control.SetValidateTicket(currentTicket, config.MaxVotes, config.TicketsUpdateTime)
	//if err != nil {
	//	log.Fatalf("createTicket to redis failed %s", err)
	//}
	// 将当前有效的票据写入 mysql
	err := control.CreateOrTicket(currentTicket)
	if err != nil {
		log.Fatalf("createTicket to mysql failed %s", err)
	}

	for range ticker.C { // 循环监听定时器的通道
		ticketMutex.Lock()                       // 在修改 currentTicket 之前加锁
		currentTicket = generateRandomString(10) // 生成一个长度为10的随机字符串作为新票据

		//// 将当前有效票据写入 redis
		//err = control.SetValidateTicket(currentTicket, config.MaxVotes, config.TicketsUpdateTime)
		//if err != nil {
		//	log.Fatalf("createTicket to redis failed %s", err)
		//}
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

// 在Redis中设置票据的使用上限
func setTicketUsageLimitInRedis(ticketID string, limit int) error {
	// 假设使用Redis客户端rdb和上下文ctx
	_, err := db.GetRedisCLi().Set(context.Background(), ticketID, limit, config.TicketsUpdateTime).Result()
	return err
}
