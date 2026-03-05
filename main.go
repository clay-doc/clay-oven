package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

/*
Args:

	-h   help:         Show help message
	-c   config:       Specify path to config file (default: clay.yaml)
	-d   docs-dir:     Specify path to documents directory (default: ./docs)
	-o   output:       Specify output directory (default: ./output)
	-fm  folder-meta:  Specify path to folder meta file (default: dir-meta.yaml)
	-nc  no-confirm:   Do not ask for confirmation before overwriting files
	-v   verbose:      Enable verbose (debug) output
	-ci  ci:           Run in CI mode (no interactive TUI, plain output, auto-confirm)
*/

// Arg defines a single CLI argument.
type Arg struct {
	ArgName  string
	ArgDesc  string
	FullArg  string
	HasValue bool
}

func main() {
	args := defineArgs()
	argsRead := os.Args[1:]
	parsed := parseArgs(argsRead, args)

	// Set global verbose flag
	if _, ok := parsed["-v"]; ok {
		Verbose = true
	}

	// Help message
	if _, ok := parsed["-h"]; ok {
		PrintBanner()
		PrintHeader("Help")
		fmt.Println()
		for _, arg := range args {
			if arg.HasValue {
				PrintKeyVal(fmt.Sprintf("%s, %s <value>", arg.ArgName, arg.FullArg), arg.ArgDesc)
			} else {
				PrintKeyVal(fmt.Sprintf("%s, %s", arg.ArgName, arg.FullArg), arg.ArgDesc)
			}
		}
		fmt.Println()
		return
	}

	// CI mode is only enabled via the explicit --ci flag,
	// or when no TTY is available (e.g. IDE run configs, piped output).
	_, ciMode := parsed["-ci"]

	if !ciMode {
		// Check if a TTY is actually available for the TUI
		if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err != nil {
			ciMode = true
		} else {
			f.Close()
		}
	}

	if ciMode {
		// CI mode: plain output, no TUI
		sink := &CISink{}
		sink.Banner()
		RunOven(parsed, sink)
		return
	}

	// TUI mode: interactive Bubble Tea interface
	sink := NewTUISink()
	model := newTUIModel(sink)

	p := tea.NewProgram(model)
	sink.SetProgram(p)

	// Run the pipeline in a background goroutine
	go RunOven(parsed, sink)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs(argsRead []string, defs []Arg) map[string]string {
	res := map[string]string{}
	lookup := map[string]Arg{}

	for _, a := range defs {
		lookup[a.ArgName] = a
		lookup[a.FullArg] = a
	}

	for i := 0; i < len(argsRead); i++ {
		s := argsRead[i]
		if def, ok := lookup[s]; ok {
			if def.HasValue {
				if i+1 >= len(argsRead) {
					PrintError(fmt.Sprintf("Missing value for %s", s))
					os.Exit(1)
				}
				res[def.ArgName] = argsRead[i+1]
				i++ // skip the value
			} else {
				res[def.ArgName] = ""
			}
		} else {
			PrintWarn(fmt.Sprintf("Unrecognized argument: %s", s))
		}
	}

	return res
}

func defineArgs() []Arg {
	return []Arg{
		{
			ArgName:  "-h",
			ArgDesc:  "Show help message",
			FullArg:  "--help",
			HasValue: false,
		},
		{
			ArgName:  "-c",
			ArgDesc:  "Specify path to config file (default: clay.yaml)",
			FullArg:  "--config",
			HasValue: true,
		},
		{
			ArgName:  "-d",
			ArgDesc:  "Specify path to documents directory (default: docs/)",
			FullArg:  "--docs-dir",
			HasValue: true,
		},
		{
			ArgName:  "-o",
			ArgDesc:  "Specify output directory (default: docs-out/)",
			FullArg:  "--output",
			HasValue: true,
		},
		{
			ArgName:  "-fm",
			ArgDesc:  "Specify path to directory meta file (default: dir-meta.yaml)",
			FullArg:  "--dir-meta",
			HasValue: true,
		},
		{
			ArgName:  "-nc",
			ArgDesc:  "Do not ask for confirmation before overwriting files",
			FullArg:  "--no-confirm",
			HasValue: false,
		},
		{
			ArgName:  "-v",
			ArgDesc:  "Enable verbose (debug) output",
			FullArg:  "--verbose",
			HasValue: false,
		},
		{
			ArgName:  "-ci",
			ArgDesc:  "Run in CI mode (plain output, no TUI, auto-confirm)",
			FullArg:  "--ci",
			HasValue: false,
		},
	}
}
