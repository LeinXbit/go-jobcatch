package config

import "time"

type AppConfig struct {
    Cities         []CityConfig
    FetchInterval  time.Duration
    RequestTimeout time.Duration
    EnableProxy    bool // 是否启用代理池
}

type CityConfig struct {
    Name string
    Code string
}

var Default = AppConfig{
    Cities: []CityConfig{
        {Name: "北京", Code: "010000"},
        {Name: "上海", Code: "020000"},
        {Name: "广州", Code: "030000"},
        {Name: "深圳", Code: "040000"},
    },
    FetchInterval:  30 * time.Second,
    RequestTimeout: 15 * time.Second,
    EnableProxy:    false, // 默认关闭，需要时改为 true
}

func GetConfig() AppConfig {
    return Default
}