package proxy

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
)

// SingleProxyClient 单代理客户端（实现 fetcher.ProxyClient 接口）
type SingleProxyClient struct {
    ProxyURL string
    Timeout  time.Duration
}

// NewSingleProxyClient 创建单代理客户端
func NewSingleProxyClient(proxyURL string, timeout time.Duration) *SingleProxyClient {
    return &SingleProxyClient{
        ProxyURL: proxyURL,
        Timeout:  timeout,
    }
}

// DoRequest 实现 fetcher.ProxyClient 接口
func (c *SingleProxyClient) DoRequest(ctx context.Context, urlStr string) ([]byte, error) {
    proxyURL, err := url.Parse(c.ProxyURL)
    if err != nil {
        return nil, fmt.Errorf("代理地址解析失败: %w", err)
    }

    transport := &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    }

    client := &http.Client{
        Timeout:   c.Timeout,
        Transport: transport,
    }

    req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("代理请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
    }

    return io.ReadAll(resp.Body)
}