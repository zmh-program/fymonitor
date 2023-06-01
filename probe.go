package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type MonitorData struct {
	URL       string            `json:"url"`
	Responses map[string]string `json:"responses"`
}

var db *sql.DB

type MonitorResponses map[string]string

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("mysql", viper.GetString("database.dsn"))
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS monitor (
			id INT AUTO_INCREMENT PRIMARY KEY,
			url VARCHAR(256) NOT NULL,
			responses JSON,
			stamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.GET("/monitor", func(c *gin.Context) {
		rows, err := db.Query("SELECT url, responses FROM monitor")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var monitors []MonitorData
		for rows.Next() {
			var url string
			var responsesJSON string
			if err := rows.Scan(&url, &responsesJSON); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Parse responses JSON
			var responses MonitorResponses
			if err := json.Unmarshal([]byte(responsesJSON), &responses); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			monitors = append(monitors, MonitorData{URL: url, Responses: responses})
		}

		c.JSON(http.StatusOK, monitors)
	})

	r.POST("/monitor", func(c *gin.Context) {
		var newMonitorData MonitorData
		if err := c.ShouldBindJSON(&newMonitorData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Insert new monitor record with empty responses
		result, err := db.Exec("INSERT INTO monitor (url, responses) VALUES (?, ?)", newMonitorData.URL, "{}")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get inserted record ID
		id, _ := result.LastInsertId()

		c.JSON(http.StatusCreated, gin.H{"id": id})
	})

	go runMonitorChecks()

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func runMonitorChecks() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		checkMonitors()
	}
}

func checkMonitors() {
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

	for rows.Next() {
		var id int
		var url string
		var responsesJSON string
		if err := rows.Scan(&id, &url, &responsesJSON); err != nil {
			fmt.Println("Error scanning monitor:", err)
			return
		}

		// Parse existing responses
		var responses MonitorResponses
		if err := json.Unmarshal([]byte(responsesJSON), &responses); err != nil {
			responses = make(MonitorResponses)
		}

		// Update responses for current hour
		responses[getCurrentHour()] = performCheck(url)

		// Marshal responses to JSON string
		resp, err := json.Marshal(responses)
		if err != nil {
			fmt.Println("Error marshaling responses:", err)
			return
		}

		// Update monitor record with new responses
		_, _ = db.Exec("UPDATE monitor SET responses = ? WHERE id = ?", resp, id)
	}
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
