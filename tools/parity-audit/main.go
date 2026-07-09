package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/parity"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	flags := flag.NewFlagSet("parity-audit", flag.ContinueOnError)
	legacyRoot := flags.String("legacy-root", "../MHCAT", "path to the legacy MHCAT repository")
	format := flags.String("format", "markdown", "output format: markdown or json")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "parity audit flag error: %v\n", err)
		return 1
	}
	legacy, err := parity.LoadLegacySlashCommands(*legacyRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parity audit legacy load error: %v\n", err)
		return 1
	}
	audit := parity.AuditSlashCommandParity(legacy, parity.CurrentGoDefinitions())
	switch *format {
	case "markdown":
		fmt.Print(parity.RenderMarkdown(audit))
	case "json":
		payload, err := parity.RenderJSON(audit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parity audit json error: %v\n", err)
			return 1
		}
		fmt.Println(string(payload))
	default:
		fmt.Fprintf(os.Stderr, "parity audit format error: format must be markdown or json\n")
		return 1
	}
	return 0
}
