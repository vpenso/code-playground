package main

import (
	"context"
	"fmt"
	"io"
	"time"
	"os"
	"os/signal"
	"sync"
	"runtime/trace"
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

// burn keeps one processor busy until ctx is cancelled. It never sleeps, sends or receives, so it
// never gives the scheduler an opening of its own accord.
func burn(ctx context.Context) {
	sum := 0
	for ctx.Err() == nil {
		for i := range 1_000_000 {
			sum += i
		}
	}
}

func main() {

	// SPINUP_TRACE names a file to record an execution trace into, for `go tool trace`. Recording
	// costs a little, so it is off unless asked for. The deferred calls run when main returns, not
	// when this block ends: defer belongs to the function, whatever braces it is written inside.
	if path := os.Getenv("SPINUP_TRACE"); path != "" {
		file, err := os.Create(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot record a trace:", err)
			os.Exit(1)
		}
		defer file.Close()
		if err := trace.Start(file); err != nil {
			fmt.Fprintln(os.Stderr, "cannot record a trace:", err)
			os.Exit(1)
		}
		defer trace.Stop()
	}

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

	spinners.Go(func() { burn(ctx) })

	render(ctx, os.Stdout, animations, updates)

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

func render(ctx context.Context, w io.Writer, a []animation, updates <-chan update) {
	fmt.Fprint(w, "\x1b[?25l")
	defer fmt.Fprint(w, "\x1b[?25h")

	current := make([]string, len(a))
	for range a {
		fmt.Println(w)
	}
	for {
		select {
		case u := <-updates:
			current[u.line] = u.frame
			fmt.Fprintf(w, "\x1b[%dA", len(a))
			for line, anim := range a {
				fmt.Fprintf(w, "\x1b[2K%s %s\n", current[line], anim.name)
			}
		case <-ctx.Done():
			return
		}
	}
}
