package fetcher

import (
    "fmt"
    "io"
    "net/http"
    "time"

    "go-catch/pkg/proxy"
)

// Job51Fetcher 51job 抓取器
type Job51Fetcher struct {
    Timeout     time.Duration
    ProxyClient *proxy.HttpClient
    UseProxy    bool
}

// NewJob51Fetcher 创建抓取器（默认直连）
func NewJob51Fetcher(timeout time.Duration) *Job51Fetcher {
    return &Job51Fetcher{
        Timeout:  timeout,
        UseProxy: false,
    }
}

// EnableProxy 启用代理池
func (f *Job51Fetcher) EnableProxy(client *proxy.HttpClient) {
    f.ProxyClient = client
    f.UseProxy = true
}

// Fetch 抓取指定城市的第几页
func (f *Job51Fetcher) Fetch(cityCode string, page int) ([]byte, error) {
    url := buildURL(cityCode, page)

    if f.UseProxy && f.ProxyClient != nil {
        return f.fetchWithProxy(url)
    }

    return f.fetchDirect(url)
}

// fetchWithProxy 使用代理池抓取
func (f *Job51Fetcher) fetchWithProxy(url string) ([]byte, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    setRequestHeaders(req)

    resp, err := f.ProxyClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("代理请求失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
    }

    return io.ReadAll(resp.Body)
}

// fetchDirect 直连抓取
func (f *Job51Fetcher) fetchDirect(url string) ([]byte, error) {
    req, err := http.NewRequest("GET", url, nil)
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

// buildURL 构建 51job 搜索 URL
func buildURL(cityCode string, page int) string {
    return fmt.Sprintf("https://search.51job.com/list/%s,000000,0000,00,9,99,Go,2,%d.html", cityCode, page)
}

// setRequestHeaders 设置通用请求头
func setRequestHeaders(req *http.Request) {
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
    req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
    req.Header.Set("Connection", "keep-alive")
}