package main

import "time"

func Ticker(interval time.Duration, handler *EventHandler) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		handler.OnTick()
	}
}
