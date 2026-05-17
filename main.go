package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go-catch/config"
    "go-catch/core"
    "go-catch/fetcher"
    "go-catch/logger"
    "go-catch/notifier"
    "go-catch/parser"
    "go-catch/storage"
    "go-catch/pkg/proxy"
    "go.uber.org/zap"
)

func main() {
    cfg := config.GetConfig()

    // 初始化 zap 日志
    err := logger.Init(logger.Config{
        Level:      cfg.LogLevel,
        Encoding:   cfg.LogEncoding,
        OutputPath: cfg.LogFile,
    })
    if err != nil {
        panic("初始化日志失败: " + err.Error())
    }
    defer logger.Sync()

    logger.Log.Info("==========================================")
    logger.Log.Info("Go岗位监控系统 v2.0 - zap日志版")
    logger.Log.Info("==========================================")

    // ========== 根据配置选择数据源 ==========
    var jobFetcher core.Fetcher
    var jobParser core.Parser

    if cfg.DataSource == config.DataSourceMock {
        jobFetcher = fetcher.NewMockFetcher(500 * time.Millisecond)
        jobParser = parser.NewMockParser()
        logger.Log.Info("使用 Mock 数据源(模拟数据)")
    } else {
        jobFetcher = fetcher.NewJob51Fetcher(cfg.RequestTimeout)
        jobParser = parser.NewJob51Parser([]string{"go", "golang"})
        logger.Log.Info("使用真实数据源(51job)")
    }
    logger.Log.Info("抓取器初始化完成")
    logger.Log.Info("解析器初始化完成")

    // 配置代理（仅对真实抓取器有效）
    if cfg.DataSource == config.DataSourceReal {
        switch cfg.ProxyMode {
        case config.ModeSingle:
            logger.Log.Info("使用单代理模式",
                zap.String("proxy_url", cfg.ProxyURL),
            )
            if realFetcher, ok := jobFetcher.(*fetcher.Job51Fetcher); ok {
                proxyClient := proxy.NewSingleProxyClient(cfg.ProxyURL, cfg.RequestTimeout)
                realFetcher.SetProxyClient(proxyClient)
                logger.Log.Info("单代理配置完成")
            }

        case config.ModePool:
            logger.Log.Info("正在初始化代理池...")

            var proxyList []*proxy.Proxy

            if cfg.ProxyPoolSize > 0 {
                source := proxy.NewFreeProxySource("https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=5000")
                proxies, err := source.Fetch()
                if err != nil || len(proxies) == 0 {
                    logger.Log.Warn("获取免费代理失败，使用默认代理",
                        zap.Error(err),
                    )
                    proxyList = []*proxy.Proxy{
                        {Host: "127.0.0.1", Port: 7890},
                    }
                } else {
                    proxyList = proxies
                    if len(proxyList) > cfg.ProxyPoolSize {
                        proxyList = proxyList[:cfg.ProxyPoolSize]
                    }
                }
            } else {
                proxyList = []*proxy.Proxy{
                    {Host: "127.0.0.1", Port: 7890},
                }
            }

            if len(proxyList) > 0 {
                proxyClient := proxy.NewProxyPoolClient(proxyList, cfg.RequestTimeout, 3)
                if proxyClient != nil {
                    if realFetcher, ok := jobFetcher.(*fetcher.Job51Fetcher); ok {
                        realFetcher.SetProxyClient(proxyClient)
                        logger.Log.Info("代理池配置完成",
                            zap.Int("proxy_count", len(proxyList)),
                        )
                    }
                } else {
                    logger.Log.Warn("代理池初始化失败，使用直连模式")
                }
            } else {
                logger.Log.Warn("没有可用代理，使用直连模式")
            }

        default:
            logger.Log.Info("使用直连模式")
        }
    }

    // ========== 创建存储层（支持 Redis + MySQL 组合） ==========
    var jobStorage core.Storage
    var mysqlStorage *storage.MySQLStorage
    var redisStorage *storage.RedisStorage

    // 1. 初始化 MySQL（如果启用）
    if cfg.DBEnabled {
        logger.Log.Info("正在连接 MySQL...")
        mysqlCfg := storage.MySQLConfig{
            Host:     cfg.DBHost,
            Port:     cfg.DBPort,
            User:     cfg.DBUser,
            Password: cfg.DBPassword,
            DBName:   cfg.DBName,
        }

        mysqlStorage, err = storage.NewMySQLStorage(mysqlCfg)
        if err != nil {
            logger.Log.Error("MySQL 连接失败，降级使用内存存储", zap.Error(err))
            mysqlStorage = nil
        } else {
            logger.Log.Info("MySQL 存储初始化完成")
        }
    }

    // 2. 初始化 Redis（如果启用）
    if cfg.Redis.Enabled {
        logger.Log.Info("正在连接 Redis...")
        redisStorage, err = storage.NewRedisStorage(cfg.Redis)
        if err != nil {
            logger.Log.Error("Redis 连接失败，将不使用 Redis 缓存", zap.Error(err))
            redisStorage = nil
        } else {
            logger.Log.Info("Redis 存储初始化完成")
        }
    }

    // 3. 组合存储层（优先级：Redis → MySQL → 内存）
    if redisStorage != nil && mysqlStorage != nil {
        // 使用 Redis 作为去重缓存，MySQL 作为持久化存储
        jobStorage = storage.NewTieredStorage(redisStorage, mysqlStorage)
        logger.Log.Info("使用分层存储(Redis 缓存 + MySQL 持久化)")
    } else if mysqlStorage != nil {
        jobStorage = mysqlStorage
        logger.Log.Info("使用 MySQL 存储")
    } else if redisStorage != nil {
        jobStorage = redisStorage
        logger.Log.Info("使用 Redis 存储（仅去重，无持久化）")
    } else {
        jobStorage = storage.NewMemoryStorage()
        logger.Log.Info("使用内存存储")
    }

    // 创建通知器
    jobNotifier := notifier.NewConsoleNotifier()
    logger.Log.Info("通知器初始化完成")

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

    // 监听系统信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        logger.Log.Info("收到中断信号，正在优雅关闭...")

        // 关闭 MySQL 连接
        if mysqlStorage != nil {
            if err := mysqlStorage.Close(); err != nil {
                logger.Log.Error("关闭 MySQL 连接失败", zap.Error(err))
            }
        }

        // 关闭 Redis 连接
        if redisStorage != nil {
            if err := redisStorage.Close(); err != nil {
                logger.Log.Error("关闭 Redis 连接失败", zap.Error(err))
            }
        }

        cancel()
    }()

    logger.Log.Info("==========================================")
    logger.Log.Info("抓取间隔",
        zap.Duration("interval", cfg.FetchInterval),
    )
    logger.Log.Info("按 Ctrl+C 停止监控")
    logger.Log.Info("数据源",
        zap.String("source", string(cfg.DataSource)),
    )
    if cfg.DataSource == config.DataSourceReal {
        logger.Log.Info("代理模式",
            zap.String("mode", string(cfg.ProxyMode)),
        )
    }
    logger.Log.Info("存储模式",
        zap.Bool("mysql", cfg.DBEnabled),
        zap.Bool("redis", cfg.Redis.Enabled),
    )
    logger.Log.Info("==========================================")

    // 启动监控
    monitor.Start(ctx)

    logger.Log.Info("程序已退出")
}