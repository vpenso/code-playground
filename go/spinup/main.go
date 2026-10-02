package main

import (
	"context"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	updates := make(chan update)

	var spinners sync.WaitGroup
	for line, a := range animations {
		spinners.Go(func() { spin(ctx, line, a, updates) })
	}

	render(ctx, animations, updates)

	spinners.Wait()
}

func spin(ctx context.Context, line int, a animation, out chan<- update) {
	ticker := time.NewTicker(a.period)
	defer ticker.Stop()
	defer fmt.Fprintln(os.Stderr, a.name, "stopped")
	for i := 0; ; i++ {
		select {
		case out <- update{line: line, frame: a.frames[i%len(a.frames)]}:
		case <-ctx.Done():
			return
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func render(ctx context.Context, animations []animation, updates <-chan update) {
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
		case <-ctx.Done():
			return
		}
	}
}
