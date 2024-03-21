package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"sync"
	"time"
)

var (
	config                 GlobalConfig  // 全局配置文件
	once                   sync.Once     // 只执行一次的代码
	MaxVotes               int           // 票据最大使用次数
	TicketsUpdateTime      time.Duration // 票据更新时间
	updateDebounceTimer    *time.Timer   // 配置更新防抖动
	TicketCacheRefreshTime time.Duration // 票数缓存刷新时间
)

const debounceDuration = 1 * time.Second

type GlobalConfig struct {
	DbConfig    DbConf    `yaml:"db" mapstructure:"db"`       // 数据库配置
	RedisConfig RedisConf `yaml:"redis" mapstructure:"redis"` // redis 配置
}

type DbConf struct {
	Host        string `yaml:"host" mapstructure:"host"`                   // 主机地址
	Port        string `yaml:"port" mapstructure:"port"`                   // 端口号
	User        string `yaml:"user" mapstructure:"user"`                   // 用户名
	Password    string `yaml:"password" mapstructure:"password"`           // 密码
	Dbname      string `yaml:"dbname" mapstructure:"dbname"`               // 数据库名
	MaxIdleConn int    `yaml:"max_idle_conn" mapstructure:"max_idle_conn"` // 最大空闲连接数
	MaxOpenConn int    `yaml:"max_open_conn" mapstructure:"max_open_conn"` // 最大打开连接数
	MaxIdleTime int64  `yaml:"max_idle_time" mapstructure:"max_idle_time"` // 连接最大空闲时间
}

// RedisConf 配置
type RedisConf struct {
	Host     string `yaml:"rhost" mapstructure:"rhost"`       // db主机地址
	Port     int    `yaml:"rport" mapstructure:"rport"`       // db端口
	DB       int    `yaml:"rdb" mapstructure:"rdb"`           // 数据库
	PassWord string `yaml:"passwd" mapstructure:"passwd"`     // 密码
	PoolSile int    `yaml:"poolsize" mapstructure:"poolsize"` // 连接池大小，即最大连接数
}

func GetGlobalConf() *GlobalConfig {
	once.Do(readConf)
	return &config
}

// 将配置文件中的信息全部加载到 全局配置文件中
func readConf() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	err := viper.ReadInConfig() // 读取配置信息
	if err != nil {
		panic("read config file err:" + err.Error())
	}
	err = viper.Unmarshal(&config) // 将配置信息反序列化填充到全局配置文件中
	if err != nil {
		panic("config file unmarshal err:" + err.Error())
	}
	log.Infof("config === %+v\n", config)

	MaxVotes = viper.GetInt("maxVotes")
	TicketsUpdateTime = viper.GetDuration("ticketUpdateTime")
	TicketCacheRefreshTime = viper.GetDuration("ticketCacheRefreshTime")
	fmt.Printf("票据最大使用次数：%d, 票据更新时间：%fs，票数缓存失效时间：%fs\n", MaxVotes, TicketsUpdateTime.Seconds(), TicketCacheRefreshTime.Seconds())
	viper.WatchConfig() //监听配置文件的变化
	viper.OnConfigChange(func(e fsnotify.Event) {
		if updateDebounceTimer != nil {
			updateDebounceTimer.Stop()
		}
		updateDebounceTimer = time.AfterFunc(debounceDuration, func() {
			viper.ReadInConfig() //重新加载
			fmt.Println(fmt.Sprintf("%s --- %s", time.Now(), "更新配置项"))
			MaxVotes = viper.GetInt("maxVotes")
			TicketsUpdateTime = viper.GetDuration("ticketUpdateTime")
		})
	})

}

func init() {
	globalConf := GetGlobalConf() // 获取全局配置文件
	fmt.Println(globalConf)
}
