package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type MonitorResponses map[string]string

func RunMonitorChecks(db *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		checkMonitors(db)
	}
}

func checkMonitors(db *sql.DB) {
	rows, err := db.Query("SELECT id, url, responses FROM monitor")
	if err != nil {
		fmt.Println("Error fetching monitors:", err)
		return
	}
	defer rows.Close()

	var wg sync.WaitGroup
	for rows.Next() {
		var id int
		var url string
		var responsesJSON string
		if err := rows.Scan(&id, &url, &responsesJSON); err != nil {
			fmt.Println("Error scanning monitor:", err)
			return
		}

		wg.Add(1)
		go func(id int, url string, responsesJSON string) {
			defer wg.Done()

			// Parse existing responses
			var responses MonitorResponses
			if err := json.Unmarshal([]byte(responsesJSON), &responses); err != nil {
				responses = make(MonitorResponses)
			}

			// Perform check
			response := performCheck(url)

			// Get current hour
			hour := getCurrentHour()

			// Update responses for current hour
			responses[hour] = response

			// Marshal responses to JSON string
			newResponsesJSON, err := json.Marshal(responses)
			if err != nil {
				fmt.Println("Error marshaling responses:", err)
				return
			}

			// Update monitor record with new responses
			_, _ = db.Exec("UPDATE monitor SET responses = ? WHERE id = ?", newResponsesJSON, id)
		}(id, url, responsesJSON)
	}

	wg.Wait()
}

func performCheck(url string) string {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error performing check:", err)
		return "-1"
	}
	defer resp.Body.Close()

	duration := time.Since(start).Milliseconds()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return fmt.Sprintf("%dms", duration)
	}

	return "-1"
}

func getCurrentHour() string {
	currentTime := time.Now()
	return currentTime.Format("15:04")
}
