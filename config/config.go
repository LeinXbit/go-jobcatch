package config

import "time"

// ProxyMode 代理模式类型
type ProxyMode string

const (
    ModeDirect ProxyMode = "direct" // 直连模式
    ModeSingle ProxyMode = "single" // 单代理模式
    ModePool   ProxyMode = "pool"   // 代理池模式
)

// AppConfig 应用配置
type AppConfig struct {
    Cities         []CityConfig
    FetchInterval  time.Duration
    RequestTimeout time.Duration

    // 代理配置
    ProxyMode     ProxyMode // 代理模式: direct, single, pool
    ProxyURL      string    // 单代理地址，如 "http://127.0.0.1:7890"
    ProxySources  []string  // 代理池源 URL 列表
    ProxyPoolSize int       // 代理池最大代理数量（0 表示使用全部）
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

    // 代理配置（默认使用代理池）
    ProxyMode:     ModeDirect,
    ProxyURL:      "http://127.0.0.1:7890",
    ProxyPoolSize: 10, // 最多使用 10 个代理
    ProxySources: []string{
        "https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=5000",
        "https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
    },
}

// GetConfig 获取配置
func GetConfig() AppConfig {
    // TODO: 从 YAML 文件读取配置（可选）
    return Default
}