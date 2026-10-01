package main

import (
	"fmt"
	"time"
)

const tick = 100 * time.Millisecond

func main() {
	frames := make(chan string)
	go moon(frames, 1)
	for frame := range frames {
		fmt.Printf("\r%s", frame)
	}
	fmt.Println()
}

func moon(out chan<- string, rounds int) {
	frames := []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	for range rounds {
		for _, frame := range frames {
			out <- frame
			time.Sleep(tick)
		}
	}
	close(out)
}

func clock(rounds int) {
	frames := []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}
	for range rounds{
		for _, frame := range frames {
			fmt.Printf("\r%s", frame)
			time.Sleep(tick)
		}
	}
}
