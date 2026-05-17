package config

import "time"

// ProxyMode 代理模式
type ProxyMode string

const (
    ModeDirect ProxyMode = "direct"
    ModeSingle ProxyMode = "single"
    ModePool   ProxyMode = "pool"
)

// DataSource 数据源类型
type DataSource string

const (
    DataSourceReal DataSource = "real" // 真实抓取（51job）
    DataSourceMock DataSource = "mock" // 模拟数据（用于测试和演示）
)

// RedisConfig Redis 配置
type RedisConfig struct {
    Enabled  bool   // 是否启用 Redis
    Host     string // Redis 地址
    Port     int    // Redis 端口
    Password string // Redis 密码（无密码留空）
    DB       int    // 数据库编号（0-15）
    TTL      int    // 缓存过期时间（秒）
}

// AppConfig 应用配置
type AppConfig struct {
    Cities         []CityConfig
    FetchInterval  time.Duration
    RequestTimeout time.Duration

    // 数据源配置
    DataSource DataSource // real 或 mock

    // 代理配置
    ProxyMode     ProxyMode
    ProxyURL      string
    ProxySources  []string
    ProxyPoolSize int

    // 数据库配置
    DBEnabled    bool
    DBHost       string
    DBPort       int
    DBUser       string
    DBPassword   string
    DBName       string

    // Redis 配置
    Redis RedisConfig

    // 日志配置
    LogLevel    string
    LogEncoding string
    LogFile     string
}

// CityConfig 城市配置
type CityConfig struct {
    Name string
    Code string
}

// Default 默认配置
var Default = AppConfig{
    Cities: []CityConfig{
        {Name: "北京", Code: "010000"},
        {Name: "上海", Code: "020000"},
        {Name: "广州", Code: "030000"},
        {Name: "深圳", Code: "040000"},
    },
    FetchInterval:  30 * time.Second,
    RequestTimeout: 15 * time.Second,

    // 数据源配置（默认使用模拟数据）
    DataSource: DataSourceReal,

    // 代理配置
    ProxyMode:     ModeDirect,
    ProxyURL:      "http://127.0.0.1:7890",
    ProxyPoolSize: 10,
    ProxySources: []string{
        "https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=5000",
    },

    // 数据库配置
    DBEnabled:  false,
    DBHost:     "127.0.0.1",
    DBPort:     3307,
    DBUser:     "root",
    DBPassword: "123456",
    DBName:     "job_monitor",

    // Redis 配置（默认关闭）
    Redis: RedisConfig{
        Enabled:  false,      // 默认关闭，需要时改为 true
        Host:     "127.0.0.1",
        Port:     6379,
        Password: "",
        DB:       0,
        TTL:      3600, // 1 小时
    },

    // 日志配置
    LogLevel:    "info",
    LogEncoding: "console",
    LogFile:     "logs/app.log",
}

// GetConfig 获取配置
func GetConfig() AppConfig {
    return Default
}