package utils

import (
	"VoteMe/db"
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestGenerateTicket(t *testing.T) {
	db.InitRedis() // 确保Redis已经初始化
	ticket1 := GetCurrentTicket()
	t.Log(ticket1)
	time.Sleep(20 * time.Second)
	ticket2 := GetCurrentTicket()
	t.Log(ticket2)

	assert.NotEqual(t, ticket1, ticket2, "Each ticket should be unique")
}

// 测试多机竞争问题
func TestConcurrencyUpdateUserVotes(t *testing.T) {
	fmt.Println(runtime.GOMAXPROCS(0))
	db.InitDB()    // 初始化数据库
	db.InitRedis() // 初始化Redis

	userName := "Bob"
	initialVotes, err := db.GetUserVotes(userName)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	votesToAdd := 1000
	wg.Add(votesToAdd)
	for i := 0; i < votesToAdd; i++ {
		go func() {
			defer wg.Done()
			err := db.UpdateUserVotesWithLock(userName)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	finalVotes, err := db.GetUserVotes(userName)
	assert.NoError(t, err)

	assert.Equal(t, initialVotes+votesToAdd, finalVotes, "User votes should accurately reflect the number of votes added in a concurrent environment")
}

func TestUpdateUserVotesExecutionTime(t *testing.T) {
	db.InitDB()                                                   // 假设DB是全局变量，这里初始化你的数据库连接
	defer db.DB.Exec("DELETE FROM users where name = 'TestUser'") // 测试完成后清理数据

	// 首先创建一个测试用户
	user := db.User{Name: "TestUser", Votes: 0}
	if err := db.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	startTime := time.Now() // 开始时间
	// 调用UpdateUserVotes函数
	err := db.UpdateUserVotesDirectSQL("TestUser")
	duration := time.Since(startTime) // 计算执行时间

	// 打印执行时间
	t.Logf("UpdateUserVotes executed in %v", duration)

	if err != nil {
		t.Errorf("UpdateUserVotes returned an error: %v", err)
	}

	// 验证投票数增加了1
	var updatedUser db.User
	if err := db.DB.Where("name = ?", "TestUser").First(&updatedUser).Error; err != nil {
		t.Fatalf("Failed to query updated user: %v", err)
	}

	if updatedUser.Votes != user.Votes+1 {
		t.Errorf("Expected votes to increase by 1, but got %d", updatedUser.Votes)
	}
}
