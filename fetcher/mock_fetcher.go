package fetcher

import (
    "context"
    "encoding/json"
    "time"

    "go-catch/model"
)

// MockFetcher 模拟抓取器
type MockFetcher struct {
    delay time.Duration
}

// NewMockFetcher 创建模拟抓取器
func NewMockFetcher(delay time.Duration) *MockFetcher {
    return &MockFetcher{delay: delay}
}

// Fetch 实现 core.Fetcher 接口
func (m *MockFetcher) Fetch(ctx context.Context, cityCode string, page int) ([]byte, error) {
    // 模拟网络延迟
    select {
    case <-time.After(m.delay):
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    // 根据页码返回不同的模拟数据
    mockJobs := m.getMockJobs(cityCode, page)
    return json.Marshal(mockJobs)
}

func (m *MockFetcher) getMockJobs(cityCode string, page int) []model.Job {
    // 城市名称映射
    cityName := map[string]string{
        "010000": "北京",
        "020000": "上海",
        "030000": "广州",
        "040000": "深圳",
    }[cityCode]
    if cityName == "" {
        cityName = "未知"
    }

    // 第一页返回数据，第二页返回空（模拟分页结束）
    if page > 1 {
        return []model.Job{}
    }

    return []model.Job{
        {
            JobID:   "mock_001_" + cityCode,
            Title:   "Go后端开发工程师",
            Company: "字节跳动",
            City:    cityName,
            Salary:  "25-40K·16薪",
            URL:     "https://example.com/job/001",
            Source:  "mock",
        },
        {
            JobID:   "mock_002_" + cityCode,
            Title:   "Golang架构师",
            Company: "腾讯",
            City:    cityName,
            Salary:  "35-50K·15薪",
            URL:     "https://example.com/job/002",
            Source:  "mock",
        },
        {
            JobID:   "mock_003_" + cityCode,
            Title:   "资深Go开发工程师",
            Company: "阿里巴巴",
            City:    cityName,
            Salary:  "30-45K·16薪",
            URL:     "https://example.com/job/003",
            Source:  "mock",
        },
        {
            JobID:   "mock_004_" + cityCode,
            Title:   "Go云原生开发",
            Company: "华为",
            City:    cityName,
            Salary:  "28-42K·14薪",
            URL:     "https://example.com/job/004",
            Source:  "mock",
        },
        {
            JobID:   "mock_005_" + cityCode,
            Title:   "后端开发工程师(Go)",
            Company: "美团",
            City:    cityName,
            Salary:  "26-38K·15薪",
            URL:     "https://example.com/job/005",
            Source:  "mock",
        },
    }
}