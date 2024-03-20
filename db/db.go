package db

import (
	"VoteMe/config"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
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
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的地方）
		logger.Config{
			SlowThreshold: time.Millisecond * 2000, // 慢SQL阈值设置为200毫秒
			LogLevel:      logger.Warn,             // 日志级别
			Colorful:      true,                    // 彩色打印
		},
	)
	// 使用gorm.Open创建数据库连接
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err) // 连接失败，记录日志并终止程序
	}
	// 在InitDB函数中添加Ticket自动迁移
	//DB.AutoMigrate(&User{}, &Ticket{})
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(mysqlConf.MaxIdleConn)                                        // 最大空闲连接
	sqlDB.SetMaxOpenConns(mysqlConf.MaxOpenConn)                                        // 最大打开连接
	sqlDB.SetConnMaxLifetime(time.Duration(mysqlConf.MaxIdleTime * int64(time.Second))) // 最大空闲时间（s）
}
func GetDB() *gorm.DB {
	dbOnce.Do(initDB)
	return db
}
