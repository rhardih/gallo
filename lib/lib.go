package lib

import (
	"fmt"
	"log"
	"os"
	"time"
)

func MustGetEnv(variable string) string {
	val, ok := os.LookupEnv(variable)
	if !ok {
		log.Fatal(fmt.Sprintf("The environment variable %s is required.", variable))
	}
	return val
}

func RunningTime(s string) (string, time.Time) {
	log.Println("Start: ", s)
	return s, time.Now()
}

func Track(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("End: ", s, "took", endTime.Sub(startTime))
}
