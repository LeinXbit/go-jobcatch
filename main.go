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

var cities = []string{"BeiJing","ShenZhen","ShangHai","GuangZhou","HangZhou"}

func main(){
	ticker := time.NewTicker(30*time.Second)
	run()

	for range ticker.C {
		run()
	}
}

func run(){
	fmt.Println("Start fetching jobs...")
	var wg sync.WaitGroup
	jobChan := make(chan interface{}, 100)

	for _, city := range cities {
		wg.Add(1)
		go fetchCity(city, jobChan, &wg)
	}

	go func() {
		wg.Wait()
		close(jobChan)
	}()

	storage.ClearNewJobs()

	for item := range jobChan {
		if job, ok := item.(parser.JobResult); ok {
			if job.Err != nil {
				fmt.Printf("Error fetching jobs for %s: %v\n", job.City, job.Err)
			}else {
				for _, j := range job.Jobs {
					if containGo(j.Title) && storage.AddIfNew(j) {
						fmt.Printf("New job found: %s at %s in %s\n", j.Title, j.Company, j.City)
						// Here you can add code to send notifications, e.g., email or push notifications
						// sendNotification(j)
					}
				}
				fmt.Printf("Finished processing jobs for %s\n, total %d\n", job.City,len(job.Jobs))
			}
		}
	}
	newJobs := storage.GetNewJobs()
	fmt.Printf("Total new jobs found: %d\n", len(newJobs))
}

func fetchCity(city string, ch chan<- interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	data, err := fetcher.Fetch(city)
	if err != nil {
		ch <- &parser.JobResult{City: city, Err: err}
		return
	}
	jobs, err := parser.ParseJobs(data)
	ch <- &parser.JobResult{City: city, Jobs: jobs, Err: err}
}

func containGo(title string) bool {
	lower := strings.ToLower(title)
	return strings.Contains(lower, "go") || strings.Contains(lower, "golang")
}