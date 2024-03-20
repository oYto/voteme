package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"time"
)

var DB *gorm.DB

// User 定义用户模型
type User struct {
	gorm.Model        // 内嵌gorm.Model，包含ID、CreatedAt、UpdatedAt等
	Name       string `gorm:"unique"` // 用户名字段，设置为唯一
	Votes      int    // 票数字段
	Version    int    // 添加版本号字段
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
	dsn := "root:123456@tcp(47.92.151.211:13306)/voteme?charset=utf8mb4&parseTime=True&loc=Local"
	// 使用gorm.Open创建数据库连接
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err) // 连接失败，记录日志并终止程序
	}

	// 在InitDB函数中添加Ticket自动迁移
	//DB.AutoMigrate(&User{}, &Ticket{})
}

// UpdateUserVotes 更新用户票数
// 接受一个用户名作为参数，将该用户的票数增加1
//
//	func UpdateUserVotes(userName string) error {
//		var user User
//		tx := DB.Where("name = ?", userName).FirstOrCreate(&user)
//		if tx.Error != nil {
//			return tx.Error // 如果操作出错，返回错误
//		}
//
//		user.Votes++                // 用户票数增加1
//		return DB.Save(&user).Error // 保存更改到数据库，如果出错返回错误
//	}

// UpdateUserVotes 5 ms，理论上来说，这种方式直接淘汰。
func UpdateUserVotes(userName string) error {
	// 构建并执行一个SQL更新语句来直接增加用户的票数
	// 这里假设用户表名为`users`，并且有`name`和`votes`列
	result := DB.Exec("UPDATE users SET votes = votes + 1 WHERE name = ?", userName)

	if result.Error != nil {
		return result.Error // 如果执行SQL语句出错，返回错误
	}

	if result.RowsAffected == 0 {
		// 如果没有更新到任何行，说明没有找到该用户名的用户，你可能需要处理这种情况
		return fmt.Errorf("no user found with name: %s", userName)
	}

	return nil // 成功更新票数
}

// UpdateUserVotesWithRetry 重试间隔和次数
func UpdateUserVotesWithRetry(userName string) error {
	var err error
	maxRetries := 10
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = UpdateUserVotesDirectSQL(userName)
		if err == nil {
			return nil // 成功，返回nil
		}
		// 这里可以根据错误类型决定是否重试
		// 例如，如果是因为版本冲突导致的更新失败，可能会希望重试
		// 如果是其他类型的错误，可能就不重试
		// 为了简化示例，这里假设所有错误都重试
		time.Sleep(time.Duration(rand.Intn(50)+10) * time.Millisecond)
	}
	return fmt.Errorf("failed to update user votes after %d attempts: %v", maxRetries, err)
}

// UpdateUserVotesOptimistically for update 6 毫秒
func UpdateUserVotesOptimistically(userName string) error {
	// 开启一个事务
	tx := DB.Begin()

	var user User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("name = ?", userName).First(&user).Error; err != nil {
		tx.Rollback() // 回滚事务
		return err
	}

	// 更新操作
	user.Votes++
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback() // 回滚事务
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// UpdateUserVotesMutex 6ms
func UpdateUserVotesMutex(userName string) error {

	var user User
	result := DB.Where("name = ?", userName).First(&user)
	if result.Error != nil {
		return result.Error
	}

	// 尝试更新记录，同时增加版本号
	result = DB.Model(&user).Where("version = ?", user.Version).Updates(User{
		Votes:   user.Votes + 1,
		Version: user.Version + 1,
	})

	if result.RowsAffected > 0 {
		// 成功更新
		return nil
	}

	// 如果没有记录被更新，说明在读取后这条记录已被其他操作更新
	// 等待一定时间后重试

	// 所有尝试都失败
	return fmt.Errorf("failed to update user votes after %d attempts due to version conflict")
}

func UpdateUserVotesDirectSQL(userName string) error {
	// SQL更新语句，同时增加votes和version字段
	sql := `UPDATE users SET votes = votes + 1, version = version + 1 WHERE name = ? AND version = (SELECT version FROM (SELECT version FROM users WHERE name = ?) AS v)`

	// 执行SQL语句
	result := DB.Exec(sql, userName, userName)

	if result.Error != nil {
		return result.Error // 如果执行SQL语句出错，返回错误
	}

	if result.RowsAffected == 0 {
		// 如果没有记录被更新，可能是因为版本号不匹配导致的，可以认为是乐观锁冲突
		return fmt.Errorf("optimistic lock conflict or no user found with name: %s", userName)
	}

	return nil // 成功更新
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
