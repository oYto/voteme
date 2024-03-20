package utils

import (
	"VoteMe/config"
	"VoteMe/control"
	"VoteMe/db"
	"VoteMe/model"
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"sync"
	"testing"
	"time"
)

// 测试多机竞争问题
func TestConcurrencyUpdateUserVotes(t *testing.T) {
	fmt.Println(runtime.GOMAXPROCS(0))
	db.GetDB()       // 初始化数据库
	db.GetRedisCLi() // 初始化Redis

	userName := "Bob"
	initialVotes, err := control.GetUserVotes(userName)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	votesToAdd := 1000
	wg.Add(votesToAdd)
	for i := 0; i < votesToAdd; i++ {
		go func() {
			defer wg.Done()
			err := control.UpdateUserVotesWithRetry(userName)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	finalVotes, err := control.GetUserVotes(userName)
	assert.NoError(t, err)

	assert.Equal(t, initialVotes+votesToAdd, finalVotes, "User votes should accurately reflect the number of votes added in a concurrent environment")
}

func TestUpdateUserVotesExecutionTime(t *testing.T) {
	defer db.GetDB().Exec("DELETE FROM users where name = 'TestUser'") // 测试完成后清理数据

	// 首先创建一个测试用户
	user := model.User{Name: "TestUser", Votes: 0}
	if err := db.GetDB().Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	startTime := time.Now() // 开始时间
	// 调用UpdateUserVotes函数
	err := control.UpdateUserVotesDirectSQL("TestUser")
	duration := time.Since(startTime) // 计算执行时间

	// 打印执行时间
	t.Logf("UpdateUserVotes executed in %v", duration)

	if err != nil {
		t.Errorf("UpdateUserVotes returned an error: %v", err)
	}

	// 验证投票数增加了1
	var updatedUser model.User
	if err := db.GetDB().Where("name = ?", "TestUser").First(&updatedUser).Error; err != nil {
		t.Fatalf("Failed to query updated user: %v", err)
	}

	if updatedUser.Votes != user.Votes+1 {
		t.Errorf("Expected votes to increase by 1, but got %d", updatedUser.Votes)
	}
}

func TestGetUserVotesPerformance(t *testing.T) {
	db.GetDB() // 初始化数据库
	fmt.Println(config.GetGlobalConf().DbConfig.MaxOpenConn)
	// 假设 "Alice" 是数据库中一个有效的用户名
	userName := "Alice"

	// 要测试的请求次数
	requests := 1000

	// 开始计时
	startTime := time.Now()

	votesToAdd := 1000
	var wg sync.WaitGroup
	wg.Add(votesToAdd)
	for i := 0; i < votesToAdd; i++ {
		go func() {
			defer wg.Done()
			_, err := control.GetUserVotes(userName)
			assert.NoError(t, err)
		}()
	}

	// 计算总耗时
	duration := time.Since(startTime)

	t.Logf("Processed %d requests in %v", requests, duration)
}
