package proxy

import (
    "fmt"
    "net/url"
)

// proxyToURL 将 Proxy 转换为 *url.URL（私有函数）
func proxyToURL(p *Proxy) *url.URL {
    if p.Username != "" && p.Password != "" {
        return &url.URL{
            Scheme: "http",
            User:   url.UserPassword(p.Username, p.Password),
            Host:   fmt.Sprintf("%s:%d", p.Host, p.Port),
        }
    }
    return &url.URL{
        Scheme: "http",
        Host:   fmt.Sprintf("%s:%d", p.Host, p.Port),
    }
}

// ProxyToURL 将代理转换为 URL（公开方法）
func ProxyToURL(proxy *Proxy) *url.URL {
    return proxyToURL(proxy)
}

// IsProxyValid 检查代理是否有效
func IsProxyValid(proxy *Proxy) bool {
    return proxy != nil && proxy.Host != "" && proxy.Port > 0 && proxy.Port < 65536
}