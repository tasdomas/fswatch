package main

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/fsnotify/fsnotify"
)

func TestOpFilter(t *testing.T) {
	c := qt.New(t)

	_, err := newOpListFilter([]string{"no", "such", "event"})
	c.Assert(err, qt.ErrorMatches, `unknown event type "no"`)

	filter, err := newOpListFilter([]string{"create", "CHMOD", "Write"})
	c.Assert(err, qt.IsNil)
	c.Assert(filter.Pass(fsnotify.Create), qt.IsTrue)
	c.Assert(filter.Pass(fsnotify.Chmod), qt.IsTrue)
	c.Assert(filter.Pass(fsnotify.Write), qt.IsTrue)
	c.Assert(filter.Pass(fsnotify.Rename), qt.IsFalse)
}

func TestOpSlice(t *testing.T) {
	c := qt.New(t)
	op := fsnotify.Create | fsnotify.Rename
	c.Assert(eventList(op), qt.DeepEquals, []string{"create", "rename"})
}
