package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var DB *gorm.DB

// User 定义用户模型
type User struct {
	gorm.Model        // 内嵌gorm.Model，包含ID、CreatedAt、UpdatedAt等
	Name       string `gorm:"unique"` // 用户名字段，设置为唯一
	Votes      int    // 票数字段
}

// Ticket 在db_manager.go中添加Ticket结构体
type Ticket struct {
	gorm.Model
	TicketID  string `gorm:"uniqueIndex"`
	Uses      int    `gorm:"default:0"`
	CreatedAt time.Time
}

// InitDB 初始化数据库连接
func InitDB() {
	var err error
	// 数据源
	dsn := "root:123456@tcp(127.0.0.1:13306)/voteme?charset=utf8mb4&parseTime=True&loc=Local"
	// 使用gorm.Open创建数据库连接
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err) // 连接失败，记录日志并终止程序
	}

	// 在InitDB函数中添加Ticket自动迁移
	DB.AutoMigrate(&User{}, &Ticket{})
}

// UpdateUserVotes 更新用户票数
// 接受一个用户名作为参数，将该用户的票数增加1
func UpdateUserVotes(userName string) error {
	var user User
	tx := DB.Where("name = ?", userName).FirstOrCreate(&user)
	if tx.Error != nil {
		return tx.Error // 如果操作出错，返回错误
	}

	user.Votes++                // 用户票数增加1
	return DB.Save(&user).Error // 保存更改到数据库，如果出错返回错误
}

// GetUserVotes 获取用户票数
// 这个函数接受一个用户名作为参数，返回该用户的当前票数
func GetUserVotes(userName string) (int, error) {
	var user User
	// 查找指定用户名的用户
	tx := DB.Where("name = ?", userName).First(&user)
	if tx.Error != nil {
		return 0, tx.Error // 如果操作出错，返回0和错误信息
	}
	return user.Votes, nil // 操作成功，返回用户的票数和nil作为错误
}

// CreateOrUpdateTicket 添加创建票据记录的函数
func CreateOrUpdateTicket(ticketID string) (*Ticket, error) {
	var ticket Ticket
	err := DB.Where("ticket_id = ?", ticketID).FirstOrCreate(&ticket, Ticket{TicketID: ticketID}).Error
	if err != nil {
		return nil, err
	}

	if ticket.Uses >= 1000000 {
		return &ticket, fmt.Errorf("ticket %s has reached its maximum usage", ticketID)
	}

	ticket.Uses += 1
	err = DB.Save(&ticket).Error
	return &ticket, err
}
