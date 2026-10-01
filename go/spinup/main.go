package main

import (
	"fmt"
	"time"
)

const tick = 100 * time.Millisecond

func main() {
	moon()
}

func moon() {
	frames := []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	for { //endless loop
		for _, frame := range frames {
			fmt.Printf("\r%s", frame)
			time.Sleep(tick)
		}

	}
}
