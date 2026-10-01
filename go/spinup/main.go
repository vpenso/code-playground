package main

import (
	"fmt"
	"time"
)

const tick = 100 * time.Millisecond

type update struct {
	line int
	frame string
}

type animation struct {
	name string
	frames []string
	period time.Duration
}

func main() {
	animations := []animation{
		{name: "moon", period: 10 * tick, frames: []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}},
		{name: "clock", period: tick, frames: []string{"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚", "🕛"}},
		{name: "dots", period: tick / 2, frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}},
	}

	updates := make(chan update)
	for line, a := range animations {
		go spin(line, a, updates)
	}
	render(animations, updates)
}

func spin(line int, a animation, out chan<- update) {
	ticker := time.NewTicker(a.period)
	defer ticker.Stop()
	for i := 0; ; i++ {
		out <- update{ line: line, frame: a.frames[i%len(a.frames)] }
		<- ticker.C
	}
}

func render(animations []animation, updates <-chan update) {
	current := make([]string, len(animations))
	for range animations {
		fmt.Println()
	}
	for u := range updates {
		current[u.line] = u.frame
		fmt.Printf("\x1b[%dA", len(animations))
		for line, a := range animations {
			fmt.Printf("\x1b[2K%s %s\n", current[line], a.name)
		}
	}
}

