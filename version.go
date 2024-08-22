package main

import (
	"fmt"
)

// VERSION the version of supervisor
const VERSION = "v0.7.3"

// VersionCommand implement the flags.Commander interface
type VersionCommand struct{}

var versionCommand VersionCommand

// Execute implement Execute() method defined in flags.Commander interface, executes the given command
func (v VersionCommand) Execute(_ []string) error {
	fmt.Println(VERSION)
	return nil
}
