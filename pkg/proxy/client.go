package proxy

import (
    "fmt"
    "io"
    "net/http"
    "time"
)

// HttpClient 带代理的 HTTP 客户端
type HttpClient struct {
    pool   *ProxyPool
    config ProxyConfig
}

// NewHttpClient 创建 HTTP 客户端
func NewHttpClient(pool *ProxyPool, config ProxyConfig) *HttpClient {
    return &HttpClient{
        pool:   pool,
        config: config,
    }
}

// Do 执行请求（自动重试和切换代理）
func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
    var lastErr error

    for i := 0; i < c.config.MaxRetry; i++ {
        proxy := c.pool.GetNext()

        transport := &http.Transport{}
        if proxy != nil {
            transport.Proxy = http.ProxyURL(proxyToURL(proxy))
        }

        client := &http.Client{
            Timeout:   c.config.RequestTimeout,
            Transport: transport,
        }

        // 复制请求
        newReq := req.Clone(req.Context())

        resp, err := client.Do(newReq)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }

        if err != nil {
            lastErr = err
        } else if resp != nil {
            lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
            resp.Body.Close()
        }

        // 等待后重试
        if i < c.config.MaxRetry-1 {
            time.Sleep(c.config.RetryInterval)
        }
    }

    return nil, fmt.Errorf("重试 %d 次后失败: %w", c.config.MaxRetry, lastErr)
}

// Get 发送 GET 请求
func (c *HttpClient) Get(url string) (*http.Response, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    return c.Do(req)
}

// GetBody 发送 GET 请求并返回 body
func (c *HttpClient) GetBody(url string) ([]byte, error) {
    resp, err := c.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
    }

    return io.ReadAll(resp.Body)
}