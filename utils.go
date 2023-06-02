package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type MonitorResponses map[string]string

func getCurrentTime() string {
	currentTime := time.Now()
	return currentTime.Format("15:04")
}

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
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var group sync.WaitGroup
	for rows.Next() {
		var id int
		var url string
		var data string
		if err := rows.Scan(&id, &url, &data); err != nil {
			fmt.Println("Error scanning monitor:", err)
			return
		}

		group.Add(1)
		go func(id int, url string, data string) {
			defer group.Done()

			// Parse existing responses
			var res MonitorResponses
			if err := json.Unmarshal([]byte(data), &res); err != nil {
				res = make(MonitorResponses)
			}

			res[getCurrentTime()] = performCheck(url)

			// Marshal responses to JSON string
			result, err := json.Marshal(res)
			if err != nil {
				fmt.Println("Error marshaling responses:", err)
				return
			}

			// Update monitor record with new responses
			_, _ = db.Exec("UPDATE monitor SET responses = ? WHERE id = ?", result, id)
		}(id, url, data)
	}

	group.Wait()
}

func performCheck(url string) string {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error performing check:", err)
		return "-1"
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(resp.Body)

	duration := time.Since(start).Milliseconds()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return fmt.Sprintf("%dms", duration)
	}

	return "-1"
}
