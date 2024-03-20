package db

import (
	"VoteMe/config"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"sync"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

// InitDB 初始化数据库连接
func initDB() {
	mysqlConf := config.GetGlobalConf().DbConfig
	var err error
	// 数据源
	dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		mysqlConf.User, mysqlConf.Password, mysqlConf.Host, mysqlConf.Port, mysqlConf.Dbname)
	// 使用gorm.Open创建数据库连接
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err) // 连接失败，记录日志并终止程序
	}
	// 在InitDB函数中添加Ticket自动迁移
	//DB.AutoMigrate(&User{}, &Ticket{})
}
func GetDB() *gorm.DB {
	dbOnce.Do(initDB)
	return db
}
