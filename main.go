package main

import (
	"os"
	"runtime"
)

const (
	applicationName = "backupboxxx"
	configFilename  = "config.json"
)

var _usersDir string

// setup os-specific values
func init() {
	switch runtime.GOOS {
	case "darwin":
		_usersDir = "/Users"
	default:
		_usersDir = "/home"
	}
}

func main() {
	run(os.Args)
}
