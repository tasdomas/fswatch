package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
)

const PathPlaceholder = "{Path}"
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

// Run runs the runners command template with the provided file path and event list.
func (c *CommandRunner) Run(ctx context.Context, path string, events []string) error {
	eventList := strings.Join(events, ",")
	cmdRaw := strings.ReplaceAll(c.tpl, PathPlaceholder, path)
	cmdRaw = strings.ReplaceAll(cmdRaw, EventsPlaceholder, eventList)

	name, args := splitCommand(cmdRaw)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	c.group.Go(func() error {
		err := cmd.Run()
		if err != nil {
			log.Printf("error executing %q: %v", cmdRaw, err)
		}
		// Hide the error.
		return nil
	})
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
