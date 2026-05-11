package proxy

import (
    "bufio"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// ProxySource 代理源接口
type ProxySource interface {
    Name() string
    Fetch() ([]*Proxy, error)
}

// FreeProxySource 免费代理源
type FreeProxySource struct {
    URL     string
    Timeout time.Duration
}

func NewFreeProxySource(url string) *FreeProxySource {
    return &FreeProxySource{
        URL:     url,
        Timeout: 10 * time.Second,
    }
}

func (s *FreeProxySource) Name() string {
    return "FreeProxySource"
}

func (s *FreeProxySource) Fetch() ([]*Proxy, error) {
    client := &http.Client{Timeout: s.Timeout}
    resp, err := client.Get(s.URL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return s.parse(string(body))
}

func (s *FreeProxySource) parse(content string) ([]*Proxy, error) {
    var proxies []*Proxy
    scanner := bufio.NewScanner(strings.NewReader(content))

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        // 支持 ip:port 格式
        parts := strings.Split(line, ":")
        if len(parts) >= 2 {
            port, err := strconv.Atoi(parts[1])
            if err == nil && port > 0 && port < 65536 {
                proxies = append(proxies, &Proxy{
                    Host: parts[0],
                    Port: port,
                })
            }
        }
    }

    return proxies, nil
}

// LocalProxySource 本地代理源（Clash/V2Ray）
type LocalProxySource struct {
    Host string
    Port int
}

func NewLocalProxySource(host string, port int) *LocalProxySource {
    return &LocalProxySource{
        Host: host,
        Port: port,
    }
}

func (s *LocalProxySource) Name() string {
    return "LocalProxySource"
}

func (s *LocalProxySource) Fetch() ([]*Proxy, error) {
    return []*Proxy{
        {
            Host: s.Host,
            Port: s.Port,
        },
    }, nil
}

// ManualProxySource 手动配置代理源
type ManualProxySource struct {
    Proxies []*Proxy
}

func NewManualProxySource(proxies []*Proxy) *ManualProxySource {
    return &ManualProxySource{Proxies: proxies}
}

func (s *ManualProxySource) Name() string {
    return "ManualProxySource"
}

func (s *ManualProxySource) Fetch() ([]*Proxy, error) {
    if len(s.Proxies) == 0 {
        return nil, fmt.Errorf("没有配置手动代理")
    }
    return s.Proxies, nil
}