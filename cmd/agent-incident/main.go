package main

import (
	"github.com/shhac/agent-incident/internal/cli"
)

var version = "dev"

func main() {
	cli.Run(version)
}
