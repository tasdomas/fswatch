package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	flag "github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

var ErrSignal = fmt.Errorf("signal received")

var helpText string = `
fswatch - watch provided path for changes and run a specified command

Usage
=====

fswatch [--events <event_list>] [--command <command>] PATH

The event list is a comma separated list of events to watch for.
Supported events are:
  - chmod
  - create
  - remove
  - rename
  - write

The path where the change was detected and the corresponding events can
be passed to the command using {Path} and {Events} placeholders, which will
be substituded with the path and the comma-separated list of events:

$ fswatch --command 'echo {Path} {Events}' ./

`[1:]

func main() {
	ctx := context.Background()

	var events []string
	var cmdTpl string
	var pth string
	var cmd string

	fs := flag.NewFlagSet("fswatch", flag.ContinueOnError)
	fs.Usage = func() {
		os.Stderr.Write([]byte(helpText))
		fs.PrintDefaults()
	}
	fs.StringSliceVarP(&events, "events", "e", nil, "Comma-separated list of events to watch.")
	fs.StringVarP(&cmdTpl, "command", "c", "echo {Path} {Events}", "Command template")
	err := fs.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		return
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

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return ListenForSignals(ctx)
	})

	watcher, err := NewWatcher(pth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start watcher: %s\n", err.Error())
		os.Exit(2)
	}
	group.Go(func() error {
		<-ctx.Done()
		err := watcher.Close()
		if err != nil {
			log.Printf("failed to stop watcher: %s\n", err.Error())
			return err
		}
		return nil
	})
	listener, err := NewListener(events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup listener: %s\n", err.Error())
		os.Exit(2)
	}
	group.Go(func() error {
		return listener.Listen(ctx, watcher.Events, watcher.Errors)
	})
	runner := NewCommandRunner(cmdTpl)
	group.Go(func() error {
		return runner.Start(ctx, listener.Chan())
	})

	err = group.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "exit reason: %s\n", err.Error())
		os.Exit(1)
	}
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

// ListenForSignals listens for interrupts.
func ListenForSignals(ctx context.Context) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ch:
		return ErrSignal
	case <-ctx.Done():
		return nil
	}
}
