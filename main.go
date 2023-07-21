package main

import (
	"os"
	"shineBlockChain/cli"
)

func main() {
	defer os.Exit(0)

	commandLine := cli.CommandLine{}
	commandLine.Run()
}
