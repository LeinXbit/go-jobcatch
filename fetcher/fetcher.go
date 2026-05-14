package fetcher

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "time"
)

// Job51Fetcher 51job 抓取器
type Job51Fetcher struct {
    Timeout  time.Duration
    UseProxy bool
    ProxyURL string
}

// NewJob51Fetcher 创建抓取器
func NewJob51Fetcher(timeout time.Duration) *Job51Fetcher {
    return &Job51Fetcher{
        Timeout:  timeout,
        UseProxy: false,
    }
}

// SetProxy 设置代理
func (f *Job51Fetcher) SetProxy(proxyURL string) {
    f.ProxyURL = proxyURL
    f.UseProxy = true
}

// Fetch 实现 core.Fetcher 接口（支持 context）
func (f *Job51Fetcher) Fetch(ctx context.Context, cityCode string, page int) ([]byte, error) {
    url := buildURL(cityCode, page)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    setRequestHeaders(req)

    client := createHTTPClient(f.Timeout, f.ProxyURL)

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
    }

    return io.ReadAll(resp.Body)
}

// buildURL 构建 URL
func buildURL(cityCode string, page int) string {
    return fmt.Sprintf("https://search.51job.com/list/%s,000000,0000,00,9,99,Go,2,%d.html", cityCode, page)
}

// setRequestHeaders 设置请求头
func setRequestHeaders(req *http.Request) {
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
    req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
    req.Header.Set("Connection", "keep-alive")
}

// createHTTPClient 创建 HTTP 客户端
func createHTTPClient(timeout time.Duration, proxyURL string) *http.Client {
    transport := &http.Transport{}
    
    if proxyURL != "" {
        // 这里可以添加代理设置
    }
    
    return &http.Client{
        Timeout:   timeout,
        Transport: transport,
    }
}