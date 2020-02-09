package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestMain_Command_NetList(t *testing.T) {
	cmd := 0
	actions["net.list"] = func(c *cli.Context) error {
		cmd++
		return nil
	}

	os.Args = []string{os.Args[0], "help"}
	main()
	assert.Equal(t, 0, cmd)

	os.Args = []string{os.Args[0], "net", "list"}
	main()
	assert.Equal(t, 1, cmd)
}

func TestMain_Command_NetStat(t *testing.T) {
	cmd := 0
	actions["net.stat"] = func(c *cli.Context) error {
		cmd++
		return nil
	}

	os.Args = []string{os.Args[0], "help"}
	main()
	assert.Equal(t, 0, cmd)

	os.Args = []string{os.Args[0], "net", "stat"}
	main()
	assert.Equal(t, 1, cmd)
}

func TestMain_Command_Version(t *testing.T) {
	cmd := 0
	actions["version"] = func(c *cli.Context) error {
		cmd++
		return nil
	}

	os.Args = []string{os.Args[0], "help"}
	main()
	assert.Equal(t, 0, cmd)

	os.Args = []string{os.Args[0], "version"}
	main()
	assert.Equal(t, 1, cmd)
}
