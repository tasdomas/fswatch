package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"

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

// CommandRunner
type CommandRunner struct {
	tpl   string
	group *errgroup.Group
}

// NewCommandRunner creates a new command runner with the specified
// command template.
func NewCommandRunner(cmd string) *CommandRunner {
	return &CommandRunner{
		tpl:   cmd,
		group: new(errgroup.Group),
	}
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
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("error executing %q: %v", cmdRaw, err)
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
