package utils

import (
	"VoteMe/control"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

// 测试将票据插入redis、能否正常获取、过期删除、达到上限后不能使用
func TestTicketUsage(t *testing.T) {
	ticketID := GetCurrentTicket()
	maxVotes, ticketUpdateTime := 200, 5*time.Second
	err := control.SetValidateTicket(ticketID, maxVotes, ticketUpdateTime)
	assert.Nil(t, err)
	t.Log(ticketID)
	//time.Sleep(ticketUpdateTime)
	//err = db.GetRedisCLi().Get(context.Background(), ticketID).Err()
	//assert.NotNil(t, err)
	for i := 0; i < maxVotes; i++ {
		flag, err := control.DecreaseUsageLimit(ticketID)
		assert.Equal(t, "true", strconv.FormatBool(flag))
		assert.Nil(t, err)

		//remaining, err := db.GetRedisCLi().Get(context.Background(), ticketID).Int()
		//assert.Nil(t, err)
		//assert.Equal(t, maxVotes-i-1, remaining, "剩余次数不匹配")
	}
	//再次减少应达到上限
	flag, err := control.DecreaseUsageLimit(ticketID)
	assert.Equal(t, "false", strconv.FormatBool(flag))
	assert.NotNil(t, err)
}

// 测试将获得票的数据插入redis是否正常，并且在过期后能否从数据库重新获取，并加载
func TestGetVotesByName(t *testing.T) {
	name := "Alice"
	votes, err := control.GetVotesByName(name)
	assert.Nil(t, err)
	assert.Equal(t, 63123, votes)
}
