package main

import (
	"gitea.com/gitea/gitea-mcp/cmd"
	"gitea.com/gitea/gitea-mcp/pkg/flag"
)

var Version = "dev"

func init() {
	flag.Version = Version
}

func main() {
	cmd.Execute()
}
