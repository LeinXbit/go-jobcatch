package proxy

import (
    "fmt"
    "time"
)

// Proxy 代理配置
type Proxy struct {
    Host      string
    Port      int
    Username  string
    Password  string
    Region    string
    LastCheck time.Time
    Success   int
    Fail      int
}

// HostPort 返回 host:port
func (p *Proxy) HostPort() string {
    return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

// String 返回代理地址字符串
func (p *Proxy) String() string {
    return p.HostPort()
}

// ProxyConfig 代理池配置
type ProxyConfig struct {
    Enabled             bool          // 是否启用代理池
    HealthCheckInterval time.Duration // 健康检查间隔
    HealthCheckTimeout  time.Duration // 健康检查超时
    MaxFailCount        int           // 最大失败次数
    RequestTimeout      time.Duration // 请求超时
    MaxRetry            int           // 最大重试次数
    RetryInterval       time.Duration // 重试间隔
}

// DefaultProxyConfig 默认配置
func DefaultProxyConfig() ProxyConfig {
    return ProxyConfig{
        Enabled:             true,
        HealthCheckInterval: 5 * time.Minute,
        HealthCheckTimeout:  10 * time.Second,
        MaxFailCount:        3,
        RequestTimeout:      15 * time.Second,
        MaxRetry:            3,
        RetryInterval:       2 * time.Second,
    }
}