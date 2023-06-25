package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// StartFilter creates a new filter that will filter events issued
// by the operation type.
func StartFilter(input <-chan fsnotify.Event, ops []string) (<-chan Trigger, error) {
	var filter opFilter = allOps{}
	if len(ops) > 0 {
		var err error
		filter, err = newOpListFilter(ops)
		if err != nil {
			return nil, err
		}
	}
	output := make(chan Trigger)
	go func() {
		for evt := range input {
			if !filter.Pass(evt.Op) {
				log.Printf("skipping event %s on path %s", evt.Op, evt.Name)
				continue
			}
			log.Printf("received event %s on path %s", evt.Op, evt.Name)
			output <- Trigger{
				Path:   evt.Name,
				Events: eventList(evt.Op),
			}
		}
	}()
	return output, nil
}

type opFilter interface {
	Pass(fsnotify.Op) bool
}

type opListFilter fsnotify.Op

func newOpListFilter(ops []string) (opListFilter, error) {
	var f fsnotify.Op
	for _, op := range ops {
		switch strings.ToLower(op) {
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
			return 0, fmt.Errorf("unknown event type %q", op)
		}
	}
	return opListFilter(f), nil
}

// Pass returns whether the event should be passed-through or filtered out.
func (f opListFilter) Pass(e fsnotify.Op) bool {
	return e&fsnotify.Op(f) > 0

}

// allOps is an event filter that allows all events.
type allOps struct{}

// Pass allows all events to go through.
func (allOps) Pass(fsnotify.Op) bool { return true }

var ops = map[fsnotify.Op]string{
	fsnotify.Chmod:  "chmod",
	fsnotify.Create: "create",
	fsnotify.Remove: "remove",
	fsnotify.Rename: "rename",
	fsnotify.Write:  "write",
}

func opList(op fsnotify.Op) []string {
	var result []string
	for evt, name := range ops {
		if op.Has(evt) {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}
