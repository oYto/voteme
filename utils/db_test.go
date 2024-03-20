package utils

import (
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
