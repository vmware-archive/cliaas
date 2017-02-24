package main

import (
	"fmt"
	"os"

	"github.com/c0-ops/cliaas/commands"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewParser(&commands.Cliaas, flags.HelpFlag)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
