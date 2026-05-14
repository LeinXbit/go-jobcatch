package config

import "time"

// ProxyMode 代理模式类型
type ProxyMode string

const (
    ModeDirect ProxyMode = "direct"
    ModeSingle ProxyMode = "single"
    ModePool   ProxyMode = "pool"
)

// AppConfig 应用配置
type AppConfig struct {
    Cities         []CityConfig
    FetchInterval  time.Duration
    RequestTimeout time.Duration

    // 代理配置
    ProxyMode     ProxyMode
    ProxyURL      string
    ProxySources  []string
    ProxyPoolSize int

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

    // 代理配置
    ProxyMode:     ModeDirect,
    ProxyURL:      "http://127.0.0.1:7890",
    ProxyPoolSize: 10,
    ProxySources: []string{
        "https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=5000",
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