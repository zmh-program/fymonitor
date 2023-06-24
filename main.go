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

	route := gin.Default()

	route.Use(func(ctx *gin.Context) {
		ctx.Set("db", db)
		ctx.Next()
	})

	// Routes
	route.GET("/monitor/:id", GetMonitorHandler)
	route.POST("/monitor", AddMonitorHandler)

	go RunMonitorChecks(db)

	if err := route.Run(":8080"); err != nil {
		panic(err)
	}
}
