package model

import (
	"gorm.io/gorm"
	"time"
)

// Ticket 在db_manager.go中添加Ticket结构体
type Ticket struct {
	gorm.Model
	TicketID  string `gorm:"uniqueIndex"`
	Uses      int    `gorm:"default:0"`
	CreatedAt time.Time
}
