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
    "go-catch/pkg/proxy"
)

func main() {
    cfg := config.GetConfig()

    logger.Info("==========================================")
    logger.Info("Go岗位监控系统 v2.0 - 低耦合架构版")
    logger.Info("==========================================")

    // 创建抓取器
    jobFetcher := fetcher.NewJob51Fetcher(cfg.RequestTimeout)
    logger.Info("✓ 抓取器初始化完成")

    // 根据配置选择代理模式
    switch cfg.ProxyMode {
    case config.ModeSingle:
        // 单代理模式
        logger.Infof("使用单代理模式: %s", cfg.ProxyURL)
        proxyClient := proxy.NewSingleProxyClient(cfg.ProxyURL, cfg.RequestTimeout)
        jobFetcher.SetProxyClient(proxyClient)
        logger.Info("✓ 单代理配置完成")

    case config.ModePool:
        // 代理池模式
        logger.Info("正在初始化代理池...")

        // 获取代理列表
        var proxyList []*proxy.Proxy

        if cfg.ProxyPoolSize > 0 {
            // 从免费源获取代理
            source := proxy.NewFreeProxySource("https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=5000")
            proxies, err := source.Fetch()
            if err != nil || len(proxies) == 0 {
                logger.Warnf("获取免费代理失败: %v，使用默认代理", err)
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
            // 手动配置代理列表
            proxyList = []*proxy.Proxy{
                {Host: "127.0.0.1", Port: 7890},
            }
        }

        if len(proxyList) > 0 {
            proxyClient := proxy.NewProxyPoolClient(proxyList, cfg.RequestTimeout, 3)
            if proxyClient != nil {
                jobFetcher.SetProxyClient(proxyClient)
                logger.Infof("✓ 代理池配置完成，共 %d 个代理", len(proxyList))
            } else {
                logger.Warn("代理池初始化失败，使用直连模式")
            }
        } else {
            logger.Warn("没有可用代理，使用直连模式")
        }

    default:
        // 直连模式
        logger.Info("使用直连模式")
    }

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

    // 监听系统信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        logger.Info("收到中断信号，正在优雅关闭...")
        cancel()
    }()

    logger.Info("==========================================")
    logger.Infof("抓取间隔: %v", cfg.FetchInterval)
    logger.Info("按 Ctrl+C 停止监控")
    logger.Infof("代理模式: %s", cfg.ProxyMode)
    logger.Info("==========================================")

    // 启动监控
    monitor.Start(ctx)

    logger.Info("程序已退出")
}