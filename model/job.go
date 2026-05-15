package model

import "time"

// Job 岗位信息模型
type Job struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    JobID     string    `gorm:"column:job_id;uniqueIndex;size:64" json:"job_id"`
    Title     string    `gorm:"column:title;size:255;not null" json:"title"`
    Company   string    `gorm:"column:company;size:255;not null" json:"company"`
    City      string    `gorm:"column:city;size:100;not null" json:"city"`
    Salary    string    `gorm:"column:salary;size:100" json:"salary"`
    URL       string    `gorm:"column:url;size:500" json:"url"`
    Source    string    `gorm:"column:source;size:50;default:51job" json:"source"`
    CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Job) TableName() string {
    return "jobs"
}