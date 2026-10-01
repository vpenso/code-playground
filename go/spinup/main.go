package main

import (
	"fmt"
	"time"
)

const tick = 100 * time.Millisecond

func main() {
	moonFrames := make(chan string)
	clockFrames := make(chan string)

	go moon(moonFrames)
	go clock(clockFrames)

	render(moonFrames, clockFrames)
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

func moon(out chan<- string) {
	frames := []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	for {
		for _, frame := range frames {
			out <- frame
			time.Sleep(10 * tick)
		}
	}
}

func clock(out chan<- string) {
	frames := []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}
	for {
		for _, frame := range frames {
			out <- frame
			time.Sleep(tick)
		}
	}
}
