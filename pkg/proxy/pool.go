package proxy

import (
    "fmt"
    "net/http"
    "net/url"
    "sync"
    "time"
)

// ProxyPool 代理池
type ProxyPool struct {
    proxies   []*Proxy
    mu        sync.RWMutex
    current   int
    config    ProxyConfig
    sources   []ProxySource
    stopChan  chan struct{}
}

// NewProxyPool 创建代理池
func NewProxyPool(config ProxyConfig, sources []ProxySource) (*ProxyPool, error) {
    pool := &ProxyPool{
        proxies:  make([]*Proxy, 0),
        config:   config,
        sources:  sources,
        stopChan: make(chan struct{}),
    }

    // 首次获取代理
    if err := pool.refresh(); err != nil {
        return nil, err
    }

    // 启动健康检查
    if config.Enabled {
        go pool.healthCheckLoop()
        go pool.refreshLoop()
    }

    return pool, nil
}

// refresh 刷新代理列表
func (p *ProxyPool) refresh() error {
    allProxies := make([]*Proxy, 0)
    seen := make(map[string]bool)

    for _, source := range p.sources {
        proxies, err := source.Fetch()
        if err != nil {
            continue
        }

        for _, proxy := range proxies {
            key := proxy.HostPort()
            if !seen[key] {
                seen[key] = true
                allProxies = append(allProxies, proxy)
            }
        }
    }

    if len(allProxies) == 0 {
        return fmt.Errorf("没有获取到任何代理")
    }

    p.mu.Lock()
    p.proxies = allProxies
    p.current = 0
    p.mu.Unlock()

    return nil
}

// refreshLoop 定期刷新代理列表
func (p *ProxyPool) refreshLoop() {
    ticker := time.NewTicker(10 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            p.refresh()
        case <-p.stopChan:
            return
        }
    }
}

// healthCheckLoop 健康检查循环
func (p *ProxyPool) healthCheckLoop() {
    ticker := time.NewTicker(p.config.HealthCheckInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            p.checkAll()
        case <-p.stopChan:
            return
        }
    }
}

// checkAll 检查所有代理
func (p *ProxyPool) checkAll() {
    p.mu.RLock()
    proxies := make([]*Proxy, len(p.proxies))
    copy(proxies, p.proxies)
    p.mu.RUnlock()

    for _, proxy := range proxies {
        if !p.checkHealth(proxy) {
            p.Remove(proxy)
        }
    }
}

// checkHealth 检查代理是否可用
func (p *ProxyPool) checkHealth(proxy *Proxy) bool {
    client := &http.Client{
        Timeout: p.config.HealthCheckTimeout,
        Transport: &http.Transport{
            Proxy: http.ProxyURL(proxyToURL(proxy)),
        },
    }

    resp, err := client.Get("https://httpbin.org/ip")
    if err != nil {
        return false
    }
    defer resp.Body.Close()

    return resp.StatusCode == 200
}

// Remove 移除代理
func (p *ProxyPool) Remove(proxy *Proxy) {
    p.mu.Lock()
    defer p.mu.Unlock()

    for i, pr := range p.proxies {
        if pr.Host == proxy.Host && pr.Port == proxy.Port {
            p.proxies = append(p.proxies[:i], p.proxies[i+1:]...)
            break
        }
    }
}

// GetNext 获取下一个可用代理（轮询）
func (p *ProxyPool) GetNext() *Proxy {
    p.mu.Lock()
    defer p.mu.Unlock()

    if len(p.proxies) == 0 {
        return nil
    }

    proxy := p.proxies[p.current]
    p.current = (p.current + 1) % len(p.proxies)
    return proxy
}

// Count 获取代理数量
func (p *ProxyPool) Count() int {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return len(p.proxies)
}

// Stop 停止代理池
func (p *ProxyPool) Stop() {
    close(p.stopChan)
}

// proxyToURL 将代理转换为 URL
func proxyToURL(proxy *Proxy) *url.URL {
    if proxy.Username != "" && proxy.Password != "" {
        return &url.URL{
            Scheme: "http",
            User:   url.UserPassword(proxy.Username, proxy.Password),
            Host:   proxy.HostPort(),
        }
    }
    return &url.URL{
        Scheme: "http",
        Host:   proxy.HostPort(),
    }
}