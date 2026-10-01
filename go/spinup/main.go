package main

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"sync"
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

	done := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		close(done)
	}()

	updates := make(chan update)

	var spinners sync.WaitGroup
	for line, a := range animations {
		spinners.Go(func() { spin(line, a, updates, done) })
	}

	render(animations, updates, done)

	spinners.Wait()
}

func spin(line int, a animation, out chan<- update, done <-chan struct{}) {
	ticker := time.NewTicker(a.period)
	defer ticker.Stop()
	defer fmt.Fprintln(os.Stderr, a.name, "stopped")
	for i := 0; ; i++ {
		select {
		case out <- update{line: line, frame: a.frames[i%len(a.frames)]}:
		case <-done:
			return
		}
		select {
		case <-ticker.C:
		case <-done:
			return
		}
	}
}

func render(animations []animation, updates <-chan update, done <-chan struct{}) {
	fmt.Print("\x1b[?25l")
	defer fmt.Print("\x1b[?25h")

	current := make([]string, len(animations))
	for range animations {
		fmt.Println()
	}
	for {
		select {
		case u := <-updates:
			current[u.line] = u.frame
			fmt.Printf("\x1b[%dA", len(animations))
			for line, a := range animations {
				fmt.Printf("\x1b[2K%s %s\n", current[line], a.name)
			}
		case <-done:
			return
		}
	}
}
