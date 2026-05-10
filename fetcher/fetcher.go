package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func FetchJobs51(cityCode string, page int) ([]byte, error) {
	url := fmt.Sprintf("https://search.51job.com/list/%s,000000,0000,00.9,99,Go,2,%d.html", cityCode, page)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err!= nil {
		return nil, err
	}
	defer resp.Body.Close()

 	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}