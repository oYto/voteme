package utils

import (
	"VoteMe/config"
	"VoteMe/db"
	"VoteMe/model"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Init 初始化
func init() {
	rand.Seed(time.Now().UnixNano())
	// 收尾工作：让 redis 中缓存的投票数，能够刷盘；将redis中缓存的东西清除
	go gracefulShutdown()
	// 生成票据
	go ticketGenerator()
	// 数据库中的信息预存到 redis 中
	getDbVotesToRedis()
	// 将redis中的数据累加到mysql中
	go syncVotesToDB()
}

// GracefulShutdown 执行最后的收尾工作
func gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	// 保证redis中新增投票能够刷盘
	fmt.Println("VotesCacheToDb......")
	time.Sleep(config.VotesCacheToDbTime)
	// 所有工作完成后，将数据库缓存清除
	fmt.Println("Clear the cache related to Voteme ......")
	time.Sleep(time.Second)
	err := deleteKeysByPattern("Voteme:*")
	if err != nil {
		fmt.Printf("Clear failed, err: %s\n", err)
	}
	fmt.Println("Clear the Ticket Table...")
	time.Sleep(time.Second)
	cleanTicketTable()
	os.Exit(0)
}

// GetDbVotesToRedis 项目启动时，自动将数据库中的用户投票数据同步到Redis
func getDbVotesToRedis() error {
	var users []model.User

	// 从数据库中查询所有用户的name和votes字段
	if err := db.GetDB().Select("name").Find(&users).Error; err != nil {
		return err
	}

	ctx := context.Background()

	// 遍历用户数据，将每个用户的投票数同步到Redis
	for _, user := range users {
		// 使用用户的votes:name作为键，votes作为值
		key := fmt.Sprintf("Voteme:votes:%s", user.Name)
		if err := db.GetRedisCLi().Set(ctx, key, user.Votes, 0).Err(); err != nil {
			return fmt.Errorf("failed to set Redis key for user %s: %v", user.Name, err)
		}
	}
	return nil
}

func syncVotesToDB() {
	ticker := time.NewTicker(config.VotesCacheToDbTime) // 每一定时间间隔刷盘一次
	defer ticker.Stop()

	for range ticker.C {
		syncVotes()
	}
}
