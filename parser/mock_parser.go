package parser

import (
    "encoding/json"

    "go-catch/model"
)

// MockParser 模拟解析器
type MockParser struct{}

// NewMockParser 创建模拟解析器
func NewMockParser() *MockParser {
    return &MockParser{}
}

// Parse 实现 core.Parser 接口
func (p *MockParser) Parse(data []byte, cityName string) ([]model.Job, error) {
    var jobs []model.Job
    if err := json.Unmarshal(data, &jobs); err != nil {
        return nil, err
    }
    return jobs, nil
}