package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type MonitorData struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Establish database connection
	db, err := sql.Open("mysql", viper.GetString("database.dsn"))
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS monitor (
		id INT AUTO_INCREMENT PRIMARY KEY,
		url VARCHAR(256) NOT NULL,
		status VARCHAR(64) NOT NULL DEFAULT 'unknown',
		stamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}

	r := gin.Default()

	r.POST("/monitor", func(c *gin.Context) {
		var newMonitorData MonitorData
		if err := c.ShouldBindJSON(&newMonitorData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("INSERT INTO monitor (url) VALUES (?)", newMonitorData.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusCreated)
	})

	r.PUT("/monitor/:id", func(c *gin.Context) {
		id := c.Param("id")
		var updatedMonitorData MonitorData
		if err := c.ShouldBindJSON(&updatedMonitorData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := db.Exec("UPDATE monitor SET status = ? WHERE id = ?", updatedMonitorData.Status, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	go func() {
		for range time.Tick(time.Minute) {
			rows, err := db.Query("SELECT id, url FROM monitor")
			if err != nil {
				log.Println("Error querying database: ", err)
				continue
			}
			for rows.Next() {
				var id int
				var url string
				if err := rows.Scan(&id, &url); err != nil {
					log.Println("Error scanning row: ", err)
					continue
				}
				resp, err := http.Get(url)
				if err != nil {
					log.Println("Error checking URL: ", err)
					db.Exec("UPDATE monitor SET status = ? WHERE id = ?", "down", id)
					continue
				}
				resp.Body.Close()
				db.Exec("UPDATE monitor SET status = ?, stamp = ? WHERE id = ?", "up", time.Now(), id)
			}
		}
	}()

	r.Run()
}
