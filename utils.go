package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"fymonitor/utils"
	"sync"
	"time"
)

type MonitorResponse map[string]int

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
	defer rows.Close()

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

			var res MonitorResponse
			if err := json.Unmarshal([]byte(data), &res); err != nil {
				res = make(MonitorResponse)
			}

			res[getCurrentTime()] = utils.PerformCheck(url, "GET", nil)

			result, err := json.Marshal(res)
			if err != nil {
				fmt.Println("Error marshaling responses:", err)
				return
			}

			_, _ = db.Exec("UPDATE monitor SET responses = ? WHERE id = ?", result, id)
		}(id, url, data)
	}

	group.Wait()
}
