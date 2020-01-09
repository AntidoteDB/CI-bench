package main

import "time"

func currentTimestamp() string {
	return time.Now().Format("20060102150405")
}
