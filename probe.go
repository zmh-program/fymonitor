package main

import (
	"fmt"
	"net/http"
	"time"
)

func benchmark() {
	var c = make(chan time.Duration, 10000)
	for i := 0; i < 10000; i++ {
		go run(url, c)
	}
	for i := range c {
		fmt.Print(i)
	}
}

func run(url string, c chan time.Duration) {
	start := time.Now()
	_, err := http.Get(url)
	if err == nil {
		c <- time.Since(start)
	}
}
