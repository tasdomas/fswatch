package main

import (
	"context"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	flag "github.com/spf13/pflag"
)

func main() {
	ctx := context.Background()

	var events []string
	var cmdTpl string
	var pth string
	var cmd string

	fs := flag.NewFlagSet("fswatch", flag.ExitOnError)
	fs.StringSliceVarP(&events, "events", "e", nil, "Comma-separated list of events to watch.")
	fs.StringVarP(&cmdTpl, "command", "c", "echo {Path} {Events}", "Command template")
	err := fs.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		panic("TODO: usage")
	}

	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "path not specified\n")
		os.Exit(2)
	}
	pth = args[0]
	if len(args) > 1 {
		cmd = args[1]
	}
	_ = cmd

	watcher, err := NewWatcher(pth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start watcher: %s\n", err.Error())
		os.Exit(2)
	}

	runner := NewCommandRunner(cmdTpl)

	listener, err := NewListener(runner, events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup listener: %s\n", err.Error())
		os.Exit(2)
	}

	listener.Listen(ctx, watcher.Events, watcher.Errors)
}

// NewWatcher creates a new watcher monitoring the provided path.
func NewWatcher(path string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(path)
	if err != nil {
		return nil, fmt.Errorf("failed to monitor path %q: %w", path, err)
	}

	return watcher, nil
}
