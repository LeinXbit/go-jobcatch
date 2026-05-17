package storage

import (
    "context"
    "fmt"
    "time"

    "go-catch/config"
    "go-catch/logger"
    "go-catch/model"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

// 重试配置常量
const (
    redisRetryCount = 2                     // 额外重试次数
    redisRetryDelay = 50 * time.Millisecond // 重试间隔
)

// RedisStorage Redis 存储实现（主要用于去重和缓存）
type RedisStorage struct {
    client  *redis.Client
    ttl     time.Duration
    newJobs []model.Job
}

// NewRedisStorage 创建 Redis 存储实例
func NewRedisStorage(cfg config.RedisConfig) (*RedisStorage, error) {
    addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: cfg.Password,
        DB:       cfg.DB,
    })

    // 测试连接
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("Redis 连接失败: %w", err)
    }

    logger.Log.Info("Redis 连接成功",
        zap.String("addr", addr),
        zap.Int("db", cfg.DB),
    )

    return &RedisStorage{
        client:  client,
        ttl:     time.Duration(cfg.TTL) * time.Second,
        newJobs: make([]model.Job, 0),
    }, nil
}

// setNXWithRetry 带重试的 SETNX 操作
func (s *RedisStorage) setNXWithRetry(key string, ttl time.Duration) (bool, error) {
    ctx := context.Background()

    for attempt := 0; attempt <= redisRetryCount; attempt++ {
        ok, err := s.client.SetNX(ctx, key, "1", ttl).Result()
        if err == nil {
            return ok, nil
        }

        if attempt == redisRetryCount {
            return false, err
        }

        time.Sleep(redisRetryDelay)
    }

    return false, fmt.Errorf("unreachable")
}

// AddIfNew 添加新岗位（使用 SETNX 实现去重，带重试）
func (s *RedisStorage) AddIfNew(job model.Job) bool {
    key := fmt.Sprintf("job:%s", job.JobID)

    // 使用带重试的 SETNX
    ok, err := s.setNXWithRetry(key, s.ttl)

    if err != nil {
        logger.Log.Error("Redis SETNX 失败（重试后仍失败）",
            zap.String("job_id", job.JobID),
            zap.String("title", job.Title),
            zap.Error(err),
        )
        return false
    }

    if ok {
        // 新岗位，记录到本次新增列表
        s.newJobs = append(s.newJobs, job)
        logger.Log.Debug("Redis 去重通过",
            zap.String("job_id", job.JobID),
            zap.String("title", job.Title),
        )
    } else {
        logger.Log.Debug("Redis 去重拦截（已存在）",
            zap.String("job_id", job.JobID),
        )
    }

    return ok
}

// GetNewJobs 获取本轮新增的岗位列表
func (s *RedisStorage) GetNewJobs() []model.Job {
    return s.newJobs
}

// ClearNewJobs 清空新增岗位列表
func (s *RedisStorage) ClearNewJobs() {
    s.newJobs = make([]model.Job, 0)
}

// Exists 检查岗位是否已存在
func (s *RedisStorage) Exists(jobID string) bool {
    ctx := context.Background()
    key := fmt.Sprintf("job:%s", jobID)

    exist, err := s.client.Exists(ctx, key).Result()
    if err != nil {
        logger.Log.Error("Redis EXISTS 失败",
            zap.String("job_id", jobID),
            zap.Error(err),
        )
        return false
    }

    return exist > 0
}

// Close 关闭 Redis 连接
func (s *RedisStorage) Close() error {
    if s.client == nil {
        return nil
    }
    return s.client.Close()
}

// ========== TieredStorage 分层存储（Redis 缓存 + MySQL 持久化） ==========

// TieredStorage 分层存储实现
// 写入：先写 Redis（去重），再写 MySQL（持久化）
// 读取：优先从 Redis 读取，未命中则从 MySQL 读取
type TieredStorage struct {
    cache   *RedisStorage // 缓存层（Redis）
    primary *MySQLStorage // 持久化层（MySQL）
    newJobs []model.Job
}

// NewTieredStorage 创建分层存储
func NewTieredStorage(cache *RedisStorage, primary *MySQLStorage) *TieredStorage {
    return &TieredStorage{
        cache:   cache,
        primary: primary,
        newJobs: make([]model.Job, 0),
    }
}

// AddIfNew 添加新岗位（先 Redis 去重，再 MySQL 持久化）
func (t *TieredStorage) AddIfNew(job model.Job) bool {
    // 1. 先通过 Redis 去重（Redis 内部已带重试）
    if !t.cache.AddIfNew(job) {
        return false // Redis 中已存在，不是新岗位
    }

    // 2. 如果是新岗位，写入 MySQL 持久化
    if t.primary != nil {
        if !t.primary.AddIfNew(job) {
            // MySQL 写入失败，记录日志但返回 true（已通过 Redis 去重）
            logger.Log.Warn("MySQL 写入失败，但 Redis 已记录",
                zap.String("job_id", job.JobID),
                zap.String("title", job.Title),
            )
        }
    }

    // 3. 记录到新增列表
    t.newJobs = append(t.newJobs, job)

    return true
}

// GetNewJobs 获取本轮新增的岗位列表
func (t *TieredStorage) GetNewJobs() []model.Job {
    return t.newJobs
}

// ClearNewJobs 清空新增岗位列表
func (t *TieredStorage) ClearNewJobs() {
    t.newJobs = make([]model.Job, 0)
    if t.cache != nil {
        t.cache.ClearNewJobs()
    }
    if t.primary != nil {
        t.primary.ClearNewJobs()
    }
}