package main

import (
	"fmt"
	"time"
)

const tick = 100 * time.Millisecond

func main() {
	moon(2)
	clock(2)
	fmt.Println()
}

func moon(rounds int) {
	frames := []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	for range rounds {
		for _, frame := range frames {
			fmt.Printf("\r%s", frame)
			time.Sleep(tick)
		}

	}
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
