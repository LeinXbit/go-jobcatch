package proxy

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "sync"
    "time"
)

// ProxyPoolClient 代理池客户端（实现 fetcher.ProxyClient 接口）
type ProxyPoolClient struct {
    proxies  []*Proxy
    current  int
    mu       sync.Mutex
    timeout  time.Duration
    retryCnt int
}

// NewProxyPoolClient 创建代理池客户端
func NewProxyPoolClient(proxies []*Proxy, timeout time.Duration, retryCnt int) *ProxyPoolClient {
    if len(proxies) == 0 {
        return nil
    }
    return &ProxyPoolClient{
        proxies:  proxies,
        timeout:  timeout,
        retryCnt: retryCnt,
        current:  0,
    }
}

// DoRequest 实现 fetcher.ProxyClient 接口（自动轮换代理）
func (c *ProxyPoolClient) DoRequest(ctx context.Context, urlStr string) ([]byte, error) {
    var lastErr error

    for i := 0; i < c.retryCnt; i++ {
        // 检查 context 是否已取消
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        proxy := c.nextProxy()
        if proxy == nil {
            continue
        }

        data, err := c.requestWithProxy(ctx, urlStr, proxy)
        if err == nil {
            return data, nil
        }

        lastErr = err

        // 等待后重试
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(time.Second):
        }
    }

    return nil, fmt.Errorf("重试 %d 次后失败: %w", c.retryCnt, lastErr)
}

// nextProxy 轮询获取下一个代理
func (c *ProxyPoolClient) nextProxy() *Proxy {
    c.mu.Lock()
    defer c.mu.Unlock()

    if len(c.proxies) == 0 {
        return nil
    }

    proxy := c.proxies[c.current]
    c.current = (c.current + 1) % len(c.proxies)
    return proxy
}

// requestWithProxy 使用指定代理发送请求
func (c *ProxyPoolClient) requestWithProxy(ctx context.Context, urlStr string, proxy *Proxy) ([]byte, error) {
    proxyURL, err := url.Parse(proxy.URL())
    if err != nil {
        return nil, err
    }

    transport := &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    }

    client := &http.Client{
        Timeout:   c.timeout,
        Transport: transport,
    }

    req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
    }

    return io.ReadAll(resp.Body)
}