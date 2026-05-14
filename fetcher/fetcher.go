package fetcher

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "time"
)

// ProxyClient 代理客户端接口（fetcher 定义，不依赖外部）
type ProxyClient interface {
    DoRequest(ctx context.Context, url string) ([]byte, error)
}

// Job51Fetcher 51job 抓取器
type Job51Fetcher struct {
    Timeout     time.Duration
    proxyClient ProxyClient // 依赖接口，不依赖具体实现
}

// NewJob51Fetcher 创建抓取器
func NewJob51Fetcher(timeout time.Duration) *Job51Fetcher {
    return &Job51Fetcher{
        Timeout:     timeout,
        proxyClient: nil,
    }
}

// SetProxyClient 设置代理客户端（支持任意实现了 ProxyClient 接口的对象）
func (f *Job51Fetcher) SetProxyClient(client ProxyClient) {
    f.proxyClient = client
}

// Fetch 实现 core.Fetcher 接口
func (f *Job51Fetcher) Fetch(ctx context.Context, cityCode string, page int) ([]byte, error) {
    url := buildURL(cityCode, page)

    // 如果设置了代理客户端，使用代理
    if f.proxyClient != nil {
        return f.proxyClient.DoRequest(ctx, url)
    }

    // 否则直连
    return f.doDirectRequest(ctx, url)
}

// doDirectRequest 直连请求
func (f *Job51Fetcher) doDirectRequest(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    setRequestHeaders(req)

    client := &http.Client{
        Timeout: f.Timeout,
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("直连请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
    }

    return io.ReadAll(resp.Body)
}

// setRequestHeaders 设置请求头
func setRequestHeaders(req *http.Request) {
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
    req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
    req.Header.Set("Connection", "keep-alive")
}

// buildURL 构建 51job 搜索 URL
func buildURL(cityCode string, page int) string {
    return fmt.Sprintf("https://search.51job.com/list/%s,000000,0000,00,9,99,Go,2,%d.html", cityCode, page)
}