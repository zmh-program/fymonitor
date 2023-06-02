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
	var data string
	err := row.Scan(&url, &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse response
	var res MonitorResponses
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	monitorData := MonitorData{URL: url, Responses: res}
	c.JSON(http.StatusOK, monitorData)
}

func AddMonitorHandler(ctx *gin.Context) {
	db := ctx.MustGet("db").(*sql.DB)

	var instance MonitorData
	if err := ctx.ShouldBindJSON(&instance); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Marshal res to JSON string
	res, err := json.Marshal(make(MonitorResponses))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert new monitor record with empty response
	result, err := db.Exec("INSERT INTO monitor (url, responses) VALUES (?, ?)", instance.URL, res)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get inserted record ID
	id, _ := result.LastInsertId()

	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}
