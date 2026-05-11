package main

import (
	"go-catch/config"
	"go-catch/logger"
	"go-catch/fetcher"
	"go-catch/parser"
	"go-catch/storage"
	"go-catch/notifier"
	"go-catch/core"
)

func main(){
	// Initialize logger
	cfg := config.GetConfig()

	// start logger
	logger.Info("go monitor v2.0")

	// Initialize components
	jobFetcher := fetcher.NewJob51Fetcher(cfg.RequestTimeout)
	logger.Info("fetcher initialized")

	keywords := []string{"go", "golang", "Go", "Golang"}
	jobParser := parser.NewJob51Parser(keywords)
	logger.Infof("parser initialized : %v", keywords)

	jobStorage := storage.NewMemoryStorage()
	logger.Info("storage initialized")

	jobNotifier := notifier.NewConsoleNotifier()
	logger.Info("notifier initialized")

	//transfrom city config to core.City
	cities := make([]core.City, len(cfg.Cities))
	for i, city := range cfg.Cities {
		cities[i] = core.City{Name: city.Name, Code: city.Code}
	}
	logger.Infof("cities initialized : %d", len(cities))
	for _, c := range cities {
		logger.Infof(" - %s (%s)", c.Name, c.Code)
	}

	// Initialize and start monitor
	monitor := core.NewMonitor(
		jobFetcher,
		jobParser,
		jobStorage,	
		jobNotifier,
		cities,
		cfg.FetchInterval,
		1, // page limit, currently set to 1 for testing
	)
	logger.Info("monitor initialized")
	logger.Infof("fetch interval: %v", cfg.FetchInterval)
	logger.Infof("press Ctrl+C to stop")

	monitor.Start()
}