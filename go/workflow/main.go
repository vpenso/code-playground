// spinup creates a local Kubernetes cluster with kind and deploys to it, showing one status line
// per step.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime/trace"
	"strings"
	"sync"
	"time"
)

const (
	tick = 100 * time.Millisecond

	// The cluster this tool owns, and the kubectl context kind creates for it. Every kubectl call
	// names the context, so the tool can never act on whichever cluster happens to be selected.
	cluster     = "spinup"
	kubeContext = "kind-" + cluster
	nodeImage   = "kindest/node:v1.36.1"
)

// step is one piece of work: a name for the display, and the work itself. run is given a function
// to report progress with, one line of text at a time.
type step struct {
	name string
	run  func(ctx context.Context, report func(text string)) error
}

func main() {
	// SPINUP_TRACE names a file to record an execution trace into, for `go tool trace`.
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

	steps := []step{
		{name: "create cluster", run: command("kind", "create", "cluster", "--name", cluster, "--image", nodeImage)},
		{name: "apply manifests", run: command("kubectl", "--context", kubeContext, "apply", "-f", "deploy")},
	}

	// Ctrl-C cancels this context, and with it every step and every command a step has started.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var names []string
	for _, s := range steps {
		names = append(names, s.name)
	}

	// The display is the only goroutine that writes to the terminal, and it outlives everything
	// that sends to it: updates is closed only after the last step has returned, so a plain send
	// on it can never be left waiting.
	updates := make(chan update)
	var display sync.WaitGroup
	display.Go(func() { render(os.Stdout, names, updates) })

	err := runSteps(ctx, steps, updates)
	close(updates)
	display.Wait()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// runSteps runs the steps one after the other and stops at the first one that fails.
func runSteps(ctx context.Context, steps []step, updates chan<- update) error {
	for line, s := range steps {
		if err := runStep(ctx, line, s, updates); err != nil {
			return err
		}
	}
	return nil
}

// runStep runs one step with a spinner beside it, and replaces the spinner with a mark when the
// step has finished.
func runStep(ctx context.Context, line int, s step, updates chan<- update) error {
	spinning, stopSpinner := context.WithCancel(ctx)
	var spinner sync.WaitGroup
	spinner.Go(func() { spin(spinning, line, working, updates) })

	started := time.Now()
	err := s.run(ctx, func(text string) { updates <- update{line: line, text: text} })

	// Cancelling asks the spinner to stop, and Wait is how we know that it has. Without the Wait a
	// last frame could still arrive after the mark below, and draw over it.
	stopSpinner()
	spinner.Wait()

	if err != nil {
		text := firstLine(err.Error())
		if ctx.Err() != nil {
			// The step did not fail by itself. It was stopped, and the error says by what.
			text = "stopped: " + text
		}
		updates <- update{line: line, frame: "✗", text: text}
		return fmt.Errorf("%s: %w", s.name, err)
	}
	took := time.Since(started).Round(100 * time.Millisecond)
	updates <- update{line: line, frame: "✓", text: took.String()}
	return nil
}

// command returns the work of a step that runs one external command. The command's output goes
// into a lineWriter and never to the terminal, which has one owner and it is not this function.
func command(name string, args ...string) func(context.Context, func(string)) error {
	return func(ctx context.Context, report func(string)) error {
		cmd := exec.CommandContext(ctx, name, args...)
		output := &lineWriter{report: report}
		cmd.Stdout = output
		cmd.Stderr = output

		// When ctx ends, ask the command to stop the way Ctrl-C would, and allow it five seconds.
		// The default is to kill it outright, and then to wait for as long as any process it
		// started keeps its output open, which can be far longer than the timeout asked for.
		cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
		cmd.WaitDelay = 5 * time.Second

		err := cmd.Run()
		if ctx.Err() != nil {
			return context.Cause(ctx)
		}
		if err != nil {
			return fmt.Errorf("%w\n%s", err, output.tail(5))
		}
		return nil
	}
}

// lineWriter is an io.Writer that reports each complete line it is given, and remembers them all.
type lineWriter struct {
	report  func(string)
	partial string
	lines   []string
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.partial += string(p)
	for {
		line, rest, found := strings.Cut(w.partial, "\n")
		if !found {
			return len(p), nil
		}
		w.partial = rest
		if line = strings.TrimSpace(line); line != "" {
			w.lines = append(w.lines, line)
			w.report(line)
		}
	}
}

// tail returns the last n lines written, for the report of a command that failed.
func (w *lineWriter) tail(n int) string {
	lines := w.lines
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return line
}
