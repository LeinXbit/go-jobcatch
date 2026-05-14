package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "go-catch/config"
    "go-catch/core"
    "go-catch/fetcher"
    "go-catch/logger"
    "go-catch/notifier"
    "go-catch/parser"
    "go-catch/storage"
)

func main() {
    cfg := config.GetConfig()

    logger.Info("==========================================")
    logger.Info("Go岗位监控系统 v2.0 - 低耦合架构版")
    logger.Info("==========================================")

    // 创建抓取器
    jobFetcher := fetcher.NewJob51Fetcher(cfg.RequestTimeout)
    logger.Info("✓ 抓取器初始化完成")

    // 创建解析器
    jobParser := parser.NewJob51Parser([]string{"go", "golang"})
    logger.Info("✓ 解析器初始化完成")

    // 创建存储器
    jobStorage := storage.NewMemoryStorage()
    logger.Info("✓ 存储器初始化完成")

    // 创建通知器
    jobNotifier := notifier.NewConsoleNotifier()
    logger.Info("✓ 通知器初始化完成")

    // 转换城市配置
    cities := make([]core.City, len(cfg.Cities))
    for i, c := range cfg.Cities {
        cities[i] = core.City{Name: c.Name, Code: c.Code}
    }

    // 创建监控器
    monitor := core.NewMonitor(
        jobFetcher,
        jobParser,
        jobStorage,
        jobNotifier,
        cities,
        cfg.FetchInterval,
        1,
    )

    // 创建可取消的 context
    ctx, cancel := context.WithCancel(context.Background())
    
    // 监听系统信号（Ctrl+C）
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        logger.Info("收到中断信号，正在优雅关闭...")
        cancel() // 取消 context，通知所有 goroutine 停止
    }()

    logger.Info("==========================================")
    logger.Infof("抓取间隔: %v", cfg.FetchInterval)
    logger.Info("按 Ctrl+C 停止监控")
    logger.Info("==========================================")

    // 启动监控（会阻塞直到 context 被取消）
    monitor.Start(ctx)
    
    logger.Info("程序已退出")
}