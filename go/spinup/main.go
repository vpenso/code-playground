package main

import (
	"fmt"
	"time"
)

const tick = 100 * time.Millisecond

var (
	moonFrames = []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	clockFrames = []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}
)

func main() {
	moon := spinner(moonFrames, 10*tick)
	clock := spinner(clockFrames, tick)
	render(moon, clock)
}

func render(moonFrames, clockFrames <-chan string) {
	var moonFrame, clockFrame string
	for {
		select {
		case moonFrame = <-moonFrames:
		case clockFrame = <-clockFrames:
	        }
		fmt.Printf("\r%s %s", moonFrame, clockFrame)
	}
}

func spinner(frames []string, period time.Duration) <-chan string {
	out := make(chan string)
	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for i := 0; ; i++ {
			out <- frames[i%len(frames)]
			<- ticker.C
		}
	}()
	return out
}

