package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	fswatch "github.com/tasdomas/fswatch"
)

func TestNewWatcher(t *testing.T) {
	c := qt.New(t)
	dir := t.TempDir()
	w, err := fswatch.NewWatcher(dir)
	c.Assert(err, qt.IsNil)
	err = w.Close()
	c.Assert(err, qt.IsNil)
}
