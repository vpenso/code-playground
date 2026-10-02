package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

// An hour of animation, tested in the time it takes to run the code. Inside synctest.Test the
// clock is fake, and it jumps forward whenever every goroutine in the test is waiting.
func TestSpinKeepsTime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Half a second past the hour, so that the deadline never falls on the same instant as a
		// tick and the number of frames is exact.
		ctx, cancel := context.WithTimeout(t.Context(), time.Hour+500*time.Millisecond)
		defer cancel()

		a := animation{period: time.Second, frames: []string{"a", "b", "c"}}
		updates := make(chan update)
		go func() {
			spin(ctx, 0, a, updates)
			close(updates)
		}()

		var got []string
		for u := range updates {
			got = append(got, u.frame)
		}

		// One frame at once, then one a second: 3601 in an hour and a half second.
		if len(got) != 3601 {
			t.Fatalf("frames in an hour at one a second = %d, want 3601", len(got))
		}
		if first := strings.Join(got[:4], ""); first != "abca" {
			t.Errorf("first four frames = %q, want %q: frames must cycle in order", first, "abca")
		}
	})
}

// A spinner has to stop when it is told to, even when nobody is receiving its frames any more.
// If it does not, every goroutine in the test ends up waiting, and synctest fails the test
// instead of hanging.
func TestSpinStopsWithNoReceiver(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		a := animation{period: time.Second, frames: []string{"a"}}

		finished := make(chan struct{})
		go func() {
			spin(ctx, 0, a, make(chan update))
			close(finished)
		}()

		cancel()
		<-finished
	})
}

// render is given a buffer in place of the terminal and two updates, and then its channel is closed.
func TestRenderDrawsEachUpdate(t *testing.T) {
	updates := make(chan update)
	var drawn bytes.Buffer

	finished := make(chan struct{})
	go func() {
		render(&drawn, []string{"one", "two"}, updates)
		close(finished)
	}()

	updates <- update{line: 0, frame: "A"}
	updates <- update{line: 1, frame: "B", text: "ready"}
	close(updates)
	<-finished

	for _, want := range []string{"A one", "B two", "ready"} {
		if !strings.Contains(drawn.String(), want) {
			t.Errorf("display never showed %q", want)
		}
	}
	if !strings.HasSuffix(drawn.String(), "\x1b[?25h") {
		t.Errorf("render returned without showing the cursor again")
	}
}

// Output arrives from a command in pieces that do not respect line ends.
func TestLineWriterReportsWholeLines(t *testing.T) {
	var reported []string
	w := &lineWriter{report: func(line string) { reported = append(reported, line) }}

	for _, piece := range []string{"first li", "ne\n  second line  \n\nthi", "rd line\nunfinished"} {
		if _, err := w.Write([]byte(piece)); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	want := "first line|second line|third line"
	if got := strings.Join(reported, "|"); got != want {
		t.Errorf("reported %q, want %q", got, want)
	}
	if got := w.tail(2); got != "second line\nthird line" {
		t.Errorf("tail(2) = %q, want the last two whole lines", got)
	}
}

// A failing step stops everything after it, and its error is the one that comes back. The steps
// here are plain functions, so the test needs no cluster and no commands.
func TestFailureStopsLaterSteps(t *testing.T) {
	boom := errors.New("boom")
	laterRan := false

	steps := []step{
		{name: "first", run: func(context.Context, func(string)) error { return nil }},
		{name: "fails", run: func(context.Context, func(string)) error { return boom }},
		{name: "later", run: func(context.Context, func(string)) error {
			laterRan = true
			return nil
		}},
	}

	// Something has to receive the updates, as the display does in the real program.
	updates := make(chan update)
	drained := make(chan struct{})
	go func() {
		for range updates {
		}
		close(drained)
	}()

	err := runSteps(t.Context(), steps, updates)
	close(updates)
	<-drained

	if !errors.Is(err, boom) {
		t.Errorf("runSteps returned %v, want the failing step's error", err)
	}
	if laterRan {
		t.Errorf("a step ran after an earlier step had failed")
	}
}
