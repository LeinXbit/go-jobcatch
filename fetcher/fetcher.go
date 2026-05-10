package fetcher

import (
	"fmt"
	"io"
	"net/http"
)

func Fetch(city string) ([]byte, error) {
	url := fmt.Sprintf("https://kenzie-acadeny.github.io/job-borad-api/jobs.json?city=%s",city)
    resp, err := http.Get(url)
	if err!=  nil{
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}