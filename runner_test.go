package main_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	fswatch "github.com/tasdomas/fswatch"
)

const outputDestEnvVar = "CMD_TEST_DST"

func TestCommandRunner(t *testing.T) {
	c := qt.New(t)

	// Setup output file.
	outputDest := filepath.Join(t.TempDir(), "output.txt")
	os.Setenv("CMD_TEST_PASSTHROUGH", "yes")
	os.Setenv(outputDestEnvVar, outputDest)
	t.Cleanup(func() {
		os.Unsetenv(outputDestEnvVar)
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Setup executable path to point to the test file.
	executablePath, err := os.Executable()
	c.Assert(err, qt.IsNil)
	cmd := fmt.Sprintf("%s -test.run TestCommandRunnerStandin -- {Path} {Events}", executablePath)

	// Setup runner.
	runner := fswatch.NewCommandRunner(cmd)
	triggerChan := make(chan fswatch.Trigger, 1)
	triggerChan <- fswatch.Trigger{
		Path:   "/tmp",
		Events: []string{"evt"},
	}

	go func() {
		err := runner.Start(ctx, triggerChan)
		c.Assert(err, qt.IsNil)
	}()
	cancel()
	out := readOutput(c, outputDest)
	c.Assert(out, qt.Equals, "/tmp evt")
}

func TestCommandRunnerStandin(t *testing.T) {
	outputDest := os.Getenv("CMD_TEST_DST")
	if outputDest == "" {
		t.SkipNow()
	}
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
		}
	}
	// Write parameters to file.
	err := os.WriteFile(
		outputDest,
		[]byte(strings.Join(args, " ")),
		0644)
	if err != nil {
		os.Exit(1)
	}
}

func readOutput(c *qt.C, fname string) string {
	output, err := os.ReadFile(fname)
	c.Assert(err, qt.IsNil)
	return string(output)
}
