package main

import (
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

    // 1. 初始化抓取器
    jobFetcher := fetcher.NewJob51Fetcher(cfg.RequestTimeout)
    logger.Info("✓ 抓取器初始化完成（51job数据源）")

    // 2. 初始化代理池（可选）
    if cfg.EnableProxy {
        logger.Info("正在初始化代理池...")

        // 配置代理源
        sources := []proxy.ProxySource{
            proxy.NewFreeProxySource("https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=5000&country=all&ssl=all&anonymity=all"),

            //proxy.NewLocalProxySource("127.0.0.1", 7890), // 本地 Clash/V2Ray
        }

        proxyConfig := proxy.DefaultProxyConfig()
        proxyPool, err := proxy.NewProxyPool(proxyConfig, sources)
        if err != nil {
            logger.Warnf("代理池初始化失败: %v，将使用直连模式", err)
        } else {
            proxyClient := proxy.NewHttpClient(proxyPool, proxyConfig)
            jobFetcher.EnableProxy(proxyClient)
            logger.Infof("✓ 代理池初始化成功，共 %d 个代理", proxyPool.Count())
        }
    } else {
        logger.Info("代理池未启用，使用直连模式")
    }

    // 3. 初始化解析器
    jobParser := parser.NewJob51Parser([]string{"go", "golang"})
    logger.Info("✓ 解析器初始化完成")

    // 4. 初始化存储器
    jobStorage := storage.NewMemoryStorage()
    logger.Info("✓ 存储器初始化完成")

    // 5. 初始化通知器
    jobNotifier := notifier.NewConsoleNotifier()
    logger.Info("✓ 通知器初始化完成")

    // 6. 转换城市配置
    cities := make([]core.City, len(cfg.Cities))
    for i, c := range cfg.Cities {
        cities[i] = core.City{Name: c.Name, Code: c.Code}
    }
    logger.Infof("✓ 加载城市配置: %d个城市", len(cities))

    // 7. 创建监控器
    monitor := core.NewMonitor(
        jobFetcher,
        jobParser,
        jobStorage,
        jobNotifier,
        cities,
        cfg.FetchInterval,
        1,
    )

    logger.Info("==========================================")
    logger.Infof("抓取间隔: %v", cfg.FetchInterval)
    logger.Info("按 Ctrl+C 停止监控")
    logger.Info("==========================================")

    // 8. 启动监控
    monitor.Start()
}