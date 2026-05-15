package storage

import (
    "fmt"
    "sync"

    "go-catch/model"
    "go-catch/logger"
    "go.uber.org/zap"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    gormlogger "gorm.io/gorm/logger"
)

// MySQLConfig MySQL 配置
type MySQLConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
}

// MySQLStorage MySQL 存储实现
type MySQLStorage struct {
    db      *gorm.DB
    seen    map[string]bool
    newJobs []model.Job
    mu      sync.RWMutex
}

// NewMySQLStorage 创建 MySQL 存储
func NewMySQLStorage(cfg MySQLConfig) (*MySQLStorage, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: gormlogger.Default.LogMode(gormlogger.Silent),
    })
    if err != nil {
        return nil, fmt.Errorf("连接数据库失败: %w", err)
    }

    if err := db.AutoMigrate(&model.Job{}); err != nil {
        return nil, fmt.Errorf("自动迁移失败: %w", err)
    }

    s := &MySQLStorage{
        db:      db,
        seen:    make(map[string]bool),
        newJobs: make([]model.Job, 0),
    }

    if err := s.loadSeenJobs(); err != nil {
        logger.Log.Warn("加载缓存失败，将继续运行", zap.Error(err))
    }

    logger.Log.Info("MySQL 存储初始化成功")
    return s, nil
}

// loadSeenJobs 加载已有岗位到缓存
func (s *MySQLStorage) loadSeenJobs() error {
    var jobs []model.Job
    if err := s.db.Select("job_id").Find(&jobs).Error; err != nil {
        return err
    }

    s.mu.Lock()
    defer s.mu.Unlock()
    for _, job := range jobs {
        s.seen[job.JobID] = true
    }

    logger.Log.Info("已加载历史岗位到缓存", zap.Int("count", len(s.seen)))
    return nil
}

// AddIfNew 添加新岗位
func (s *MySQLStorage) AddIfNew(job model.Job) bool {
    s.mu.RLock()
    exists := s.seen[job.JobID]
    s.mu.RUnlock()

    if exists {
        return false
    }

    result := s.db.Create(&job)
    if result.Error != nil {
        logger.Log.Error("保存岗位失败",
            zap.String("job_id", job.JobID),
            zap.String("title", job.Title),
            zap.Error(result.Error),
        )
        return false
    }

    s.mu.Lock()
    s.seen[job.JobID] = true
    s.newJobs = append(s.newJobs, job)
    s.mu.Unlock()

    logger.Log.Debug("新增岗位已保存",
        zap.String("job_id", job.JobID),
        zap.String("title", job.Title),
    )
    return true
}

// GetNewJobs 获取新增岗位
func (s *MySQLStorage) GetNewJobs() []model.Job {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.newJobs
}

// ClearNewJobs 清空新增列表
func (s *MySQLStorage) ClearNewJobs() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.newJobs = make([]model.Job, 0)
}

// Close 关闭数据库连接
func (s *MySQLStorage) Close() error {
    if s.db == nil {
        return nil
    }
    sqlDB, err := s.db.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}