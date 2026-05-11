package main

import (
	"go-catch/config"
	"go-catch/logger"
	"go-catch/model"
	"go-catch/fetcher"
	"go-catch/parser"
	"go-catch/storage"
	"sync"
	"time"
	"strings"
)

func main(){
	cfg := config.GetConfig()

	logger.Info("Go Job Catcher started...")
	logger.Infof("Target cities:%d ", len(cfg.Cities))
	for _, city := range cfg.Cities {
		logger.Infof("%s (%s)", city.Name, city.Code)
	}
	logger.Infof("Fetching every %v seconds...", cfg.FetchInterval)
	logger.Info("Press Ctrl+C to stop.")

	ticker := time.NewTicker(cfg.FetchInterval)
	run()	

	for range ticker.C {
		run()
	}
}

func run(){
	cfg := config.GetConfig()

	logger.Infof("\nStart fetching jobs...\n", time.Now().Format("2006-01-02 15:04:05"))
	var wg sync.WaitGroup
	jobChan := make(chan parser.JobResult, 100)

	for _, city := range cfg.Cities {
		wg.Add(1)
		go fetchCity51(city.Name, city.Code, jobChan, &wg)
	}

	go func() {
		wg.Wait()
		close(jobChan)
	}()

	storage.ClearNewJobs()

	for result := range jobChan {
		if result.Err != nil {
			logger.Errorf("Error fetching %s: %v\n", result.City, result.Err)
			continue
		}

		for _, job := range result.Jobs {
			if storage.AddIfNew(job) {
				logger.Infof("New job found: [%s] %s at %s in %s\n", job.ID, job.Title, job.Company, job.City)
				// Here you can add code to send notifications, e.g., email or push notifications
				// sendNotification(job)
			}
		}
		logger.Infof("Finished processing %s, found %d jobs\n", result.City, len(result.Jobs))
	}

	newJobs := storage.GetNewJobs()
	if len(newJobs) > 0 {
		logger.Infof("Total new jobs found: %d\n", len(newJobs))
	}else {
		logger.Info("No new jobs found.")
	}
}

func fetchCity51(cityName, cityCode string, ch chan<- parser.JobResult, wg *sync.WaitGroup) {
	defer wg.Done()

	cfg := config.GetConfig()

	// 只抓取第一页，后续可以增加分页抓取
	data, err := fetcher.FetchJobs51(cityCode, 1, cfg.RequestTimeout)
	if err != nil {
		ch <- parser.JobResult{City: cityName, Err: err}
		return
	}

	jobs, err := parser.ParseJobs51(data, cityName)
	if err != nil {
		ch <- parser.JobResult{City: cityName, Err: err}
		return
	}

	var goJobs []model.Job
	for _, jobs := range jobs {
		if containGo(jobs.Title) {
			goJobs = append(goJobs, jobs)
		}
	}

	ch <- parser.JobResult{City: cityName, Jobs: goJobs, Err: err}
}

func containGo(title string) bool {
	lower := strings.ToLower(title)
	return strings.Contains(lower, "go") || strings.Contains(lower, "golang")
}