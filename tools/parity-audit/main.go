package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/parity"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("parity-audit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	legacyRoot := flags.String("legacy-root", "../MHCAT", "path to the legacy MHCAT repository")
	format := flags.String("format", "markdown", "output format: markdown or json")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintf(stderr, "parity audit flag error: %v\n", err)
		return 1
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "parity audit flag error: unexpected positional arguments: %s\n", strings.Join(flags.Args(), " "))
		return 1
	}
	legacy, err := parity.LoadLegacySlashCommands(*legacyRoot)
	if err != nil {
		fmt.Fprintf(stderr, "parity audit legacy load error: %v\n", err)
		return 1
	}
	audit := parity.AuditSlashCommandParity(legacy, parity.CurrentGoDefinitions())
	switch *format {
	case "markdown":
		fmt.Fprint(stdout, parity.RenderMarkdown(audit))
	case "json":
		payload, err := parity.RenderJSON(audit)
		if err != nil {
			fmt.Fprintf(stderr, "parity audit json error: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(payload))
	default:
		fmt.Fprintf(stderr, "parity audit format error: format must be markdown or json\n")
		return 1
	}
	if auditFailed(audit) {
		fmt.Fprintln(stderr, "parity audit failed: command definitions contain drift, gaps, duplicates, or parser errors")
		return 2
	}
	return 0
}

func auditFailed(audit parity.CommandAudit) bool {
	return len(audit.CommandsWithDrift) != 0 ||
		len(audit.MissingGoDefinitions) != 0 ||
		len(audit.ExtraGoDefinitions) != 0 ||
		len(audit.DuplicateLegacyNames) != 0 ||
		len(audit.DuplicateGoNames) != 0 ||
		audit.LegacyParseErrorCount != 0
}
