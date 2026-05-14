package core

import (
    "context"
    "sync"
    "time"

    "go-catch/logger"
    "go-catch/model"
    "go.uber.org/zap"
)

// Fetcher 抓取器接口
type Fetcher interface {
    Fetch(ctx context.Context, cityCode string, page int) ([]byte, error)
}

// Parser 解析器接口
type Parser interface {
    Parse(data []byte, cityName string) ([]model.Job, error)
}

// Storage 存储器接口
type Storage interface {
    AddIfNew(job model.Job) bool
    GetNewJobs() []model.Job
    ClearNewJobs()
}

// Notifier 通知器接口
type Notifier interface {
    NotifyNewJob(job model.Job)
    NotifyBatch(jobs []model.Job)
}

// City 城市配置
type City struct {
    Name string
    Code string
}

// Monitor 监控器
type Monitor struct {
    fetcher   Fetcher
    parser    Parser
    storage   Storage
    notifier  Notifier
    cities    []City
    interval  time.Duration
    pageLimit int
}

// NewMonitor 创建监控器
func NewMonitor(
    fetcher Fetcher,
    parser Parser,
    storage Storage,
    notifier Notifier,
    cities []City,
    interval time.Duration,
    pageLimit int,
) *Monitor {
    return &Monitor{
        fetcher:   fetcher,
        parser:    parser,
        storage:   storage,
        notifier:  notifier,
        cities:    cities,
        interval:  interval,
        pageLimit: pageLimit,
    }
}

// Start 启动监控（支持 context 取消）
func (m *Monitor) Start(ctx context.Context) {
    logger.Log.Info("Starting job monitor...")
    logger.Log.Info("Monitoring cities", zap.Int("count", len(m.cities)))
    for _, c := range m.cities {
        logger.Log.Info("City configured",
            zap.String("name", c.Name),
            zap.String("code", c.Code),
        )
    }

    ticker := time.NewTicker(m.interval)

    // 立即执行一次
    m.run(ctx)

    for {
        select {
        case <-ticker.C:
            m.run(ctx)
        case <-ctx.Done():
            logger.Log.Info("收到取消信号，监控器正在关闭...")
            ticker.Stop()
            logger.Log.Info("监控器已关闭")
            return
        }
    }
}

// run 执行一次抓取周期
func (m *Monitor) run(ctx context.Context) {
    logger.Log.Info("Checking for new jobs",
        zap.String("time", time.Now().Format("2006-01-02 15:04:05")),
    )

    var wg sync.WaitGroup
    jobChan := make(chan []model.Job, 100)

    for _, city := range m.cities {
        wg.Add(1)
        go m.fetchCity(ctx, city, jobChan, &wg)
    }

    go func() {
        wg.Wait()
        close(jobChan)
    }()

    m.storage.ClearNewJobs()

    totalJobs := 0
    for jobs := range jobChan {
        for _, job := range jobs {
            if m.storage.AddIfNew(job) {
                m.notifier.NotifyNewJob(job)
            }
        }
        totalJobs += len(jobs)
    }

    newJobs := m.storage.GetNewJobs()
    if len(newJobs) > 0 {
        logger.Log.Info("Fetch cycle completed",
            zap.Int("total_jobs", totalJobs),
            zap.Int("new_jobs", len(newJobs)),
        )
    } else {
        logger.Log.Info("No new jobs found",
            zap.Int("total_jobs", totalJobs),
        )
    }
}

// fetchCity 抓取单个城市（支持 context 超时）
func (m *Monitor) fetchCity(ctx context.Context, city City, ch chan<- []model.Job, wg *sync.WaitGroup) {
    defer wg.Done()

    var allJobs []model.Job

    for page := 1; page <= m.pageLimit; page++ {
        // 检查 context 是否已取消
        select {
        case <-ctx.Done():
            logger.Log.Warn("Task cancelled, stopping fetch",
                zap.String("city", city.Name),
            )
            return
        default:
        }

        // 创建带超时的 context（单次请求超时 30 秒）
        reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

        data, err := m.fetcher.Fetch(reqCtx, city.Code, page)
        cancel() // 立即释放资源

        if err != nil {
            logger.Log.Error("Failed to fetch data",
                zap.String("city", city.Name),
                zap.Int("page", page),
                zap.Error(err),
            )
            break
        }

        jobs, err := m.parser.Parse(data, city.Name)
        if err != nil {
            logger.Log.Error("Failed to parse data",
                zap.String("city", city.Name),
                zap.Int("page", page),
                zap.Error(err),
            )
            continue
        }

        if len(jobs) == 0 {
            break
        }

        allJobs = append(allJobs, jobs...)
        logger.Log.Info("Fetched jobs",
            zap.String("city", city.Name),
            zap.Int("page", page),
            zap.Int("count", len(jobs)),
        )

        // 避免请求过快
        time.Sleep(500 * time.Millisecond)
    }

    if len(allJobs) > 0 {
        logger.Log.Info("Total jobs fetched for city",
            zap.String("city", city.Name),
            zap.Int("total", len(allJobs)),
        )
        ch <- allJobs
    }
}