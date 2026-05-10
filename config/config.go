package config

import (
	"time"
)

type AppConfig struct {
	Cities         []CityConfig
	FetchInterval  time.Duration
	RequestTimeout time.Duration
}

type CityConfig struct {
	Name string
	Code string
}

var Default = AppConfig{
	Cities: []CityConfig{
		{Name:"BeiJing", Code: "010000"},
		{Name:"ShangHai", Code: "020000"},
		{Name:"GuangZhou", Code: "030000"},
		{Name:"ShenZhen", Code: "040000"},
	},
	FetchInterval:  30 * time.Second,
	RequestTimeout: 10 * time.Second,
}

func GetConfig() AppConfig {
	return Default
}