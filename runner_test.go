package main_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	fswatch "github.com/tasdomas/fswatch"
)

const outputDestEnvVar = "CMD_TEST_DST"

func TestRunCommand(t *testing.T) {
	c := qt.New(t)

	// Setup output file.
	outputDest := filepath.Join(t.TempDir(), "output.txt")
	os.Setenv("CMD_TEST_PASSTHROUGH", "yes")
	os.Setenv(outputDestEnvVar, outputDest)
	t.Cleanup(func() {
		os.Unsetenv(outputDestEnvVar)
	})

	ctx := context.Background()
	cmdName, err := os.Executable()
	c.Assert(err, qt.IsNil)

	cmdArgs := []string{"-test.run", "TestCommandRunnerStandin", "--", "some_path"}
	err = fswatch.RunCmd(ctx, nil, cmdName, cmdArgs...)
	c.Assert(err, qt.IsNil)

	out := readOutput(c, outputDest)
	c.Assert(out, qt.Equals, "some_path")
}

// TestCommandRunnerStandin is a dummy test function that is executed as an external command
// by the tests.
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
