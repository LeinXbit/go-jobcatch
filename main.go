package main

import (
	"strings"
	"fmt"
	"sync"
	"go-catch/fetcher"
	"go-catch/parser"
	"go-catch/storage"
	"time"
)

var cities = []struct{
	Name string
	Code string
}{
	{"BeiJing", "010000"},
	{"ShangHai", "020000"},
	{"GuangZhou", "030000"},
	{"ShenZhen", "040000"},
}

func main(){
	fmt.Println("Go Job Catcher started...")
	fmt.Println("Fetch from 51job")
	fmt.Println("Target cities: BeiJing, ShangHai, GuangZhou, ShenZhen")
	fmt.Println("Fetching every 30 seconds...")
	fmt.Println("Press Ctrl+C to stop.")

	ticker := time.NewTicker(30 * time.Second)
	run()

	for range ticker.C {
		run()
	}
}

func run(){
	fmt.Println("\nStart fetching jobs...\n", time.Now().Format("2006-01-02 15:04:05"))
	var wg sync.WaitGroup
	jobChan := make(chan parser.JobResult, 100)

	for _, city := range cities {
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
			fmt.Printf("Error fetching %s: %v\n", result.City, result.Err)
			continue
		}

		for _, job := range result.Jobs {
			if storage.AddIfNew(job) {
				fmt.Printf("New job found: [%s] %s at %s in %s\n", job.ID, job.Title, job.Company, job.City)
				// Here you can add code to send notifications, e.g., email or push notifications
				// sendNotification(job)
			}
		}
		fmt.Printf("Finished processing %s, found %d jobs\n", result.City, len(result.Jobs))
	}

	newJobs := storage.GetNewJobs()
	if len(newJobs) > 0 {
		fmt.Printf("Total new jobs found: %d\n", len(newJobs))
	}else {
		fmt.Println("No new jobs found.")
	}
}

func fetchCity51(cityName, cityCode string, ch chan<- parser.JobResult, wg *sync.WaitGroup) {
	defer wg.Done()
	data, err := fetcher.FetchJobs51(cityCode, 1)
	if err != nil {
		ch <- parser.JobResult{City: cityName, Err: err}
		return
	}
	jobs, err := parser.ParseJobs51(data, cityName)
	ch <- parser.JobResult{City: cityName, Jobs: jobs, Err: err}
}

func containGo(title string) bool {
	lower := strings.ToLower(title)
	return strings.Contains(lower, "go") || strings.Contains(lower, "golang")
}