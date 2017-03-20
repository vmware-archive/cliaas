package main

import (
	"log"

	flags "github.com/jessevdk/go-flags"
	"github.com/pivotal-cf/cliaas/commands"
)

func main() {
	parser := flags.NewParser(&commands.Cliaas, flags.HelpFlag)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}
