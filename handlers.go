package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MonitorData struct {
	URL       string            `json:"url"`
	Responses map[string]string `json:"responses"`
}

func GetMonitorHandler(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	id := c.Param("id")

	row := db.QueryRow("SELECT url, responses FROM monitor WHERE id = ?", id)

	var url string
	var responsesJSON string
	err := row.Scan(&url, &responsesJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse responses JSON
	var responses MonitorResponses
	if err := json.Unmarshal([]byte(responsesJSON), &responses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	monitorData := MonitorData{URL: url, Responses: responses}
	c.JSON(http.StatusOK, monitorData)
}

func AddMonitorHandler(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	var newMonitorData MonitorData
	if err := c.ShouldBindJSON(&newMonitorData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initialize responses
	responses := make(MonitorResponses)

	// Marshal responses to JSON string
	newResponsesJSON, err := json.Marshal(responses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert new monitor record with empty responses
	result, err := db.Exec("INSERT INTO monitor (url, responses) VALUES (?, ?)", newMonitorData.URL, newResponsesJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get inserted record ID
	id, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
