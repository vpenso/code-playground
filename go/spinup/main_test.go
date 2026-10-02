package main

import (
	"bytes"
	"context"
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

		a := animation{name: "test", period: time.Second, frames: []string{"a", "b", "c"}}
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
		a := animation{name: "test", period: time.Second, frames: []string{"a"}}

		finished := make(chan struct{})
		go func() {
			spin(ctx, 0, a, make(chan update))
			close(finished)
		}()

		cancel()
		<-finished
	})
}

// render is given a buffer in place of the terminal, two updates, and then told to stop.
func TestRenderDrawsEachUpdate(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	animations := []animation{{name: "one"}, {name: "two"}}
	updates := make(chan update)
	var drawn bytes.Buffer

	finished := make(chan struct{})
	go func() {
		render(ctx, &drawn, animations, updates)
		close(finished)
	}()

	updates <- update{line: 0, frame: "A"}
	updates <- update{line: 1, frame: "B"}
	cancel()
	<-finished

	for _, want := range []string{"A one", "B two"} {
		if !strings.Contains(drawn.String(), want) {
			t.Errorf("display never showed %q", want)
		}
	}
	if !strings.HasSuffix(drawn.String(), "\x1b[?25h") {
		t.Errorf("render returned without showing the cursor again")
	}
}
