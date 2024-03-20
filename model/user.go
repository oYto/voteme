package model

import "gorm.io/gorm"

// User 定义用户模型
type User struct {
	gorm.Model        // 内嵌gorm.Model，包含ID、CreatedAt、UpdatedAt等
	Name       string `gorm:"unique"` // 用户名字段，设置为唯一
	Votes      int    // 票数字段
	Version    int    // 添加版本号字段
}
