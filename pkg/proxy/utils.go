package proxy

import (
    "net/url"
)

// ProxyToURL 将代理转换为 URL（公开方法）
func ProxyToURL(proxy *Proxy) *url.URL {
    return proxyToURL(proxy)
}

// IsProxyValid 检查代理是否有效
func IsProxyValid(proxy *Proxy) bool {
    return proxy != nil && proxy.Host != "" && proxy.Port > 0 && proxy.Port < 65536
}