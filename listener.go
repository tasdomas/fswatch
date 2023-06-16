package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// NewListener creates a  new listener that will respond to the provided
// set of events.
func NewListener(events []string) (*Listener, error) {
	var filter eventFilter = allEvents{}
	if len(events) > 0 {
		var err error
		filter, err = newEventListFilter(events)
		if err != nil {
			return nil, err
		}
	}
	return &Listener{
		eventFilter: filter,
		output:      make(chan Trigger),
	}, nil
}

// Listener listens for operation events emitted by fsnotify.Watcher.
type Listener struct {
	eventFilter eventFilter
	output      chan Trigger
}

// Chan returns the output channel for the listener's filtered events.
func (l Listener) Chan() <-chan Trigger {
	return l.output
}

// Listen starts processing the provided event and error channels.
func (l Listener) Listen(ctx context.Context, events <-chan fsnotify.Event, errors <-chan error) error {
	for {
		select {
		case evt := <-events:
			if !l.eventFilter.Pass(evt.Op) {
				log.Printf("skipping event %s on path %s", evt.Op, evt.Name)
				continue
			}
			log.Printf("received event %s on path %s", evt.Op, evt.Name)
			l.output <- Trigger{
				Path:   evt.Name,
				Events: eventList(evt.Op),
			}
		case err := <-errors:
			log.Printf("fsnotify error: %v", err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type eventFilter interface {
	Pass(fsnotify.Op) bool
}

type eventListFilter fsnotify.Op

func newEventListFilter(events []string) (eventListFilter, error) {
	var f fsnotify.Op
	for _, evt := range events {
		switch strings.ToLower(evt) {
		case "create":
			f = f | fsnotify.Create
		case "write":
			f = f | fsnotify.Write
		case "remove":
			f = f | fsnotify.Remove
		case "rename":
			f = f | fsnotify.Rename
		case "chmod":
			f = f | fsnotify.Chmod
		default:
			return 0, fmt.Errorf("unknown event type %q", evt)
		}
	}
	return eventListFilter(f), nil
}

// Pass returns whether the event should be passed-through or filtered out.
func (f eventListFilter) Pass(e fsnotify.Op) bool {
	return e&fsnotify.Op(f) > 0

}

// allEvents is an event filter that allows all events.
type allEvents struct{}

// Pass allows all events to go through.
func (allEvents) Pass(fsnotify.Op) bool { return true }

var ops = map[fsnotify.Op]string{
	fsnotify.Chmod:  "chmod",
	fsnotify.Create: "create",
	fsnotify.Remove: "remove",
	fsnotify.Rename: "rename",
	fsnotify.Write:  "write",
}

func eventList(op fsnotify.Op) []string {
	var result []string
	for evt, name := range ops {
		if op.Has(evt) {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}
