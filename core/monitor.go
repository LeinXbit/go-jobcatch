package core

import (
	"go-catch/model"
	"go-catch/logger"
	"sync"
	"time"
)

type Fetcher interface {
	Fetch(cityCode string, page int) ([]byte, error)
}

type Parser interface {
	Parse(data []byte, cityName string) ([]model.Job, error)
}

type Storage interface {
	AddIfNew(job model.Job) bool
	GetNewJobs() []model.Job
	ClearNewJobs()
}

type Notifier interface {
	NotifyNewJob(job model.Job)
	NotifyBatch(jobs []model.Job)
}

type City struct {
	Name string
	Code string
}

type Monitor struct {
	fetcher   Fetcher
	parser    Parser
	storage   Storage
	notifier  Notifier
	cities    []City
	interval  time.Duration
	pageLimit int
}

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

func (m *Monitor) Start() {
	logger.Info("Starting job monitor...")
	logger.Infof("Monitoring cities: %d", len(m.cities))
	for _, c := range m.cities {
		logger.Infof(" - %s (%s)", c.Name, c.Code)
	}

	ticker := time.NewTicker(m.interval)
	m.run()

	for range ticker.C {
		m.run()
	}
}

func (m *Monitor) run() {
	logger.Infof("Checking for new jobs at %s...", time.Now().Format("2006-01-02 15:04:05"))

	var wg sync.WaitGroup
	jobChan := make(chan []model.Job, 100)

	for _, city := range m.cities {
		wg.Add(1)
		go m.fetchCity(city, jobChan, &wg)
	}

	go func(){
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

	logger.Infof("Finished checking. Total jobs found: %d, new jobs: %d", totalJobs, len(m.storage.GetNewJobs()))
}

func (m *Monitor) fetchCity(city City, ch chan<- []model.Job, wg *sync.WaitGroup) {
	defer wg.Done()

	var allJobs []model.Job

	for page := 1; page <= m.pageLimit; page++ {
		data, err := m.fetcher.Fetch(city.Code, page)
		if err != nil {
			logger.Errorf("Failed to fetch data for %s (page %d): %v", city.Name, page, err)
			break
		}

		jobs, err := m.parser.Parse(data, city.Name)
		if err != nil {
			logger.Errorf("Failed to parse data for %s (page %d): %v", city.Name, page, err)
			continue
		}

		if len(jobs) == 0 {
			break
		}

		allJobs = append(allJobs, jobs...)
		logger.Infof("Fetched %d jobs for %s (page %d)", len(jobs), city.Name, page)

		//avoid hitting the server too hard
		time.Sleep(500 * time.Millisecond)
	}

	if len(allJobs) > 0 {
		logger.Infof("Total %d jobs fetched for %s", len(allJobs), city.Name)
		ch <- allJobs
	}
}