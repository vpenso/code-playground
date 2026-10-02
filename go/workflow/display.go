package main

import (
	"context"
	"fmt"
	"io"
	"time"
)

// animation is what a spinner draws, and how fast.
type animation struct {
	frames []string
	period time.Duration
}

// working is the spinner shown beside a step that is running.
var working = animation{period: tick, frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}}

// update changes one line of the display. An empty frame or text leaves that part of the line as
// it was, so a spinner can change the frame while the step beside it changes the text.
type update struct {
	line  int
	frame string
	text  string
}

// spin sends the frames of one animation until ctx is cancelled. Both places where it waits also
// wait on ctx.Done(): a spinner stuck in a send that nobody will ever receive is a goroutine leaked.
func spin(ctx context.Context, line int, a animation, out chan<- update) {
	ticker := time.NewTicker(a.period)
	defer ticker.Stop()
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

// render owns the terminal. It draws one line per name and redraws them all on every update, until
// updates is closed. It hides the cursor while it draws, and the deferred call shows it again
// however render comes to return.
func render(w io.Writer, names []string, updates <-chan update) {
	fmt.Fprint(w, "\x1b[?25l")
	defer fmt.Fprint(w, "\x1b[?25h")

	frames := make([]string, len(names))
	texts := make([]string, len(names))
	for i := range names {
		frames[i] = "·"
		fmt.Fprintf(w, "%s %s\n", frames[i], names[i])
	}
	for u := range updates {
		if u.frame != "" {
			frames[u.line] = u.frame
		}
		if u.text != "" {
			texts[u.line] = u.text
		}
		fmt.Fprintf(w, "\x1b[%dA", len(names))
		for i, name := range names {
			fmt.Fprintf(w, "\x1b[2K%s %-16s %s\n", frames[i], name, texts[i])
		}
	}
}
