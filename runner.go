package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const PathPlaceholder = "{Path}"
const EventsPlaceholder = "{Events}"

var RunCmd = runCmdImpl

// RunnerOptions are used to optionally configure
// the command runner.
type RunnerOption func(*CommandRunner)

// WithTimeout sets a timeout for commands executed by the runner.
func WithTimeout(timeout time.Duration) RunnerOption {
	return func(c *CommandRunner) {
		c.timeout = timeout
	}
}

// WithLimitedConcurrent limits the number of concurrently running commands.
func WithLimitedConcurrent(limit int) RunnerOption {
	return func(c *CommandRunner) {
		c.group.SetLimit(limit)
	}
}

// CommandRunner
type CommandRunner struct {
	tpl   string
	group *errgroup.Group

	timeout time.Duration
}

// NewCommandRunner creates a new command runner with the specified
// command template.
func NewCommandRunner(cmd string, options ...RunnerOption) *CommandRunner {
	c := &CommandRunner{
		tpl:   cmd,
		group: new(errgroup.Group),
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// Run runs the runners command template with the provided file path and event list.
func (c *CommandRunner) Run(ctx context.Context, path string, events []string) error {
	eventList := strings.Join(events, ",")
	cmdRaw := strings.ReplaceAll(c.tpl, PathPlaceholder, path)
	cmdRaw = strings.ReplaceAll(cmdRaw, EventsPlaceholder, eventList)

	name, args := splitCommand(cmdRaw)
	var ctxDone context.CancelFunc
	if c.timeout != 0 {
		ctx, ctxDone = context.WithDeadline(ctx, time.Now().Add(c.timeout))
	}
	c.group.Go(func() error {
		err := RunCmd(ctx, ctxDone, name, args...)
		if err != nil {
			log.Printf("error executing %q: %v", name, err)
		}
		// Hide the error.
		return nil
	})
	return nil
}

func runCmdImpl(ctx context.Context, ctxDone context.CancelFunc, name string, args ...string) error {
	if ctxDone != nil {
		defer ctxDone()
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command error: %w", err)
	}
	return nil
}

// Wait for all commands to exit.
func (c *CommandRunner) Wait() {
	err := c.group.Wait()
	if err != nil {
		log.Printf("error waiting for subprocesses to exit: %v", err)
	}
}

func splitCommand(cmd string) (name string, args []string) {
	var quoted bool
	cmdParts := strings.FieldsFunc(cmd, func(r rune) bool {
		if r == '"' || r == '\'' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})
	if len(cmdParts) == 0 {
		return "", nil
	}
	return cmdParts[0], cmdParts[1:]
}
