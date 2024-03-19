package utils

import (
	"VoteMe/db"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
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
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	db.InitDB()    // 初始化数据库
	db.InitRedis() // 初始化Redis

	userName := "Bob"
	initialVotes, err := db.GetUserVotes(userName)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	votesToAdd := 1000
	wg.Add(votesToAdd)
	a, b := 0, 0
	start := time.Now()
	for i := 0; i < votesToAdd; i++ {
		go func() {
			defer wg.Done()
			err := db.UpdateUserVotesWithLock(userName)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
	end := time.Since(start)
	t.Logf("lock:%d, unlock:%d, time:%d\n", a, b, end)
	finalVotes, err := db.GetUserVotes(userName)
	assert.NoError(t, err)

	assert.Equal(t, initialVotes+votesToAdd, finalVotes, "User votes should accurately reflect the number of votes added in a concurrent environment")
}
