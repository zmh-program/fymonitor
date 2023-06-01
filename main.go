package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	db := ConnectDatabase()
	defer db.Close()

	r := gin.Default()

	// Middleware to set up db connection
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Routes
	r.GET("/monitor", GetMonitorsHandler)
	r.POST("/monitor", AddMonitorHandler)

	go RunMonitorChecks(db)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
