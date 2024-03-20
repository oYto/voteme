package utils

import (
	"VoteMe/config"
	"math/rand"
	"sync"
	"time"
)

var (
	currentTicket string     // 存储当前有效的票据
	ticketMutex   sync.Mutex // ticketMutex是一个互斥锁，用于控制对currentTicket变量的并发访问
)

// Init 初始化票据生成器
func Init() {
	rand.Seed(time.Now().UnixNano())
	go ticketGenerator()
}

// ticketGenerator是一个票据生成器，每20秒生成一个新的随机票据
// todo 这里为了方便测试，设置了20秒，后续改为需求中的2s
func ticketGenerator() {
	currentTicket = generateRandomString(10)
	ticker := time.NewTicker(config.TicketsUpdateTime)
	for range ticker.C { // 循环监听定时器的通道
		ticketMutex.Lock()                       // 在修改currentTicket之前加锁
		currentTicket = generateRandomString(10) // 生成一个长度为10的随机字符串作为新票据
		ticketMutex.Unlock()                     // 修改完成后解锁
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
// 该函数使用互斥锁来确保并发访问时的安全性
func GetCurrentTicket() string {
	//ticketMutex.Lock()         // 在读取currentTicket之前加锁
	//defer ticketMutex.Unlock() // 确保在函数返回前解锁
	return currentTicket // 返回当前有效的票据
}
