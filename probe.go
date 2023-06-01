package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
	"log"
	"net/http"
	"sync"
	"time"
)

var websites = []string{
	"https://www.google.com",
	"https://www.yahoo.com",
	"https://www.bing.com",
}

type Result struct {
	Website   string
	Status    string
	Timestamp string
}

func monitor(wg *sync.WaitGroup, db *sql.DB, url string) {
	defer wg.Done()
	for {
		status := "down"
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error for %s: %s", url, err)
		} else if resp.StatusCode == 200 {
			status = "up"
		}

		_, err = db.Prepare("INSERT INTO monitor_results(website, status, timestamp) values(?,?,?)")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s is %s\n", url, status)

		time.Sleep(5 * time.Second)
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, _ := db.Query("SELECT website, status, timestamp FROM monitor_results ORDER BY timestamp DESC")
	defer rows.Close()

	results := make([]*Result, 0)

	for rows.Next() {
		result := new(Result)
		err := rows.Scan(&result.Website, &result.Status, &result.Timestamp)
		if err != nil {
			return
		}
		results = append(results, result)
	}

	err := json.NewEncoder(w).Encode(results)
	if err != nil {
		return
	}
}

func main() {
	db, _ := sql.Open("sqlite3", "./monitor.db")
	defer db.Close()

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS monitor_results (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		website TEXT NOT NULL,
		status TEXT NOT NULL,
		timestamp DATETIME NOT NULL
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	var group sync.WaitGroup

	for _, website := range websites {
		group.Add(1)
		go monitor(&group, db, website)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, db)
	})

	handler := cors.Default().Handler(mux)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
