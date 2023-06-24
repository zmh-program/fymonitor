package utils

import (
	"net/http"
	"time"
)

func PerformCheck(url string, method string, headers map[string]string) int {
	start := time.Now()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return -1
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return int(time.Since(start).Milliseconds())
	}
	return -1
}
