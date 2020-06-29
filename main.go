package main

import (
	"math/rand"
	"time"
	gallo "gallo/app"
)

func main() {
	// Initialize RNG
	rand.Seed(time.Now().Unix())

	app := gallo.Application{Addr: ":8080"}
	app.Run()
}
