package db

import (
	"VoteMe/control"
	"VoteMe/utils"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// 测试将票据插入redis、能否正常获取、过期删除、达到上限后不能使用
func TestTicketUsage(t *testing.T) {
	ticketID := utils.GetCurrentTicket()
	maxVotes, ticketUpdateTime := 200, 5*time.Second
	err := control.SetValidateTicket(ticketID, maxVotes, ticketUpdateTime)
	assert.Nil(t, err)
	t.Log(ticketID)
	//time.Sleep(ticketUpdateTime)
	//err = db.GetRedisCLi().Get(context.Background(), ticketID).Err()
	//assert.NotNil(t, err)
	for i := 0; i < maxVotes; i++ {
		err := control.DecreaseUsageLimit(ticketID)
		assert.Nil(t, err)

		//remaining, err := db.GetRedisCLi().Get(context.Background(), ticketID).Int()
		//assert.Nil(t, err)
		//assert.Equal(t, maxVotes-i-1, remaining, "剩余次数不匹配")
	}
	//再次减少应达到上限
	err = control.DecreaseUsageLimit(ticketID)
	assert.NotNil(t, err)
}

// 测试将获得票的数据插入redis是否正常，并且在过期后能否从数据库重新获取，并加载
func TestGetVotesByName(t *testing.T) {
	name := "Alice"
	for i := 0; i < 1000; i++ {
		votes, err := control.GetVotesByName(name)
		assert.Nil(t, err)
		assert.Equal(t, 106803, votes)
	}
}

func TestVoteForUser(t *testing.T) {
	name := "Alice"
	n := 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			control.VoteForUserRedis(name)
		}()
	}
	wg.Wait()
}

//func TestGenerateRandomHash(t *testing.T) {
//	hash, err := GenerateRandomHash(config.TicketLen)
//	assert.Nil(t, err)
//	t.Log(hash)
//}

//func TestDeleteKeysByPattern(t *testing.T) {
//	deleteKeysByPattern("Voteme:*")
//}

//func TestGetAllUserNames(t *testing.T) {
//	names, err := GetAllUserNames()
//	assert.Nil(t, err)
//	assert.Equal(t, []string{"Alice", "Bob"}, names)
//}
//
//func TestGetDbVotesToRedis(t *testing.T) {
//	err := GetDbVotesToRedis()
//	assert.Nil(t, err)
//
//}
