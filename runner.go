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

// Trigger represents an event consumed by the runner.
type Trigger struct {
	Path   string
	Events []string
}

// PathPlaceholder is replaced with the path of the affected file
// in the command template.
const PathPlaceholder = "{Path}"

// EventsPlaceholder is replaced with a comma separated list of
// event names.
const EventsPlaceholder = "{Events}"

var RunCmd = runCmdImpl

// RunnerOptions are used to optionally configure
// the command runner.
type RunnerOption func(*CommandRunner)

// WithCommandTimeout sets a timeout for commands executed by the runner.
func WithCommandTimeout(timeout time.Duration) RunnerOption {
	return func(c *CommandRunner) {
		c.timeout = timeout
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

// Start starts listening for events on the trigger channel.
func (c *CommandRunner) Start(ctx context.Context, triggerCh <-chan Trigger) error {
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			// Wait for any running commands to finish.
			groupErr := c.group.Wait()
			if groupErr != nil {
				err = groupErr
			}
			if err != nil {
				return err
			}
			return nil
		case trigger := <-triggerCh:
			c.group.Go(
				func() error {
					c.run(ctx, trigger.Path, trigger.Events)
					return nil
				})
		}
	}
}

// run runs the runners command template with the provided file path and event list.
func (c *CommandRunner) run(ctx context.Context, path string, events []string) {
	eventList := strings.Join(events, ",")
	cmdRaw := strings.ReplaceAll(c.tpl, PathPlaceholder, path)
	cmdRaw = strings.ReplaceAll(cmdRaw, EventsPlaceholder, eventList)

	name, args := splitCommand(cmdRaw)
	var ctxDone context.CancelFunc
	if c.timeout != 0 {
		ctx, ctxDone = context.WithDeadline(ctx, time.Now().Add(c.timeout))
	}
	err := RunCmd(ctx, ctxDone, name, args...)
	if err != nil {
		log.Printf("error executing %q: %v", name, err)
	}
}

// Wait for all commands to exit.
func (c *CommandRunner) Wait() {
	err := c.group.Wait()
	if err != nil {
		log.Printf("error waiting for spawned commands: %v", err)
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
