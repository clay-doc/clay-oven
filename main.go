package main

import (
	"fmt"
	"os"
	//tea "github.com/charmbracelet/bubbletea"
)

/*
Args:

	-h help: Show help message
	-c config: Specify path to config file (default: clay.yaml)
	-d docs-dir: Specify path to documents directory (default: ./docs)
	-o output: Specify output directory (default: ./output)
	-f force: Overwrite existing files
	-fm folder-meta: Specify path to folder icons file (default: dir-meta.yaml)
	-nc no-confirm: Do not ask for confirmation before overwriting files
*/

type Arg struct {
	ArgName  string
	ArgDesc  string
	FullArg  string
	HasValue bool
}

func main() {
	fmt.Printf("\n  ______   __     __  ________  __    __ \n /      \\ |  \\   |  \\|        \\|  \\  |  \\\n|  $$$$$$\\| $$   | $$| $$$$$$$$| $$\\ | $$\n| $$  | $$| $$   | $$| $$__    | $$$\\| $$\n| $$  | $$ \\$$\\ /  $$| $$  \\   | $$$$\\ $$\n| $$  | $$  \\$$\\  $$ | $$$$$   | $$\\$$ $$\n| $$__/ $$   \\$$ $$  | $$_____ | $$ \\$$$$\n \\$$    $$    \\$$$   | $$     \\| $$  \\$$$\n  \\$$$$$$      \\$     \\$$$$$$$$ \\$$   \\$$\n                                         \n")

	args := defineArgs()

	argsRead := getArgs()
	parsed := parseArgs(argsRead, args)

	// Help message
	if _, ok := parsed["-h"]; ok {
		fmt.Println("Help for clay-oven:")
		for _, arg := range args {
			if arg.HasValue {
				fmt.Printf("  %s, %s <value>: %s\n", arg.ArgName, arg.FullArg, arg.ArgDesc)
			} else {
				fmt.Printf("  %s, %s: %s\n", arg.ArgName, arg.FullArg, arg.ArgDesc)
			}
		}
		return
	}

	RunOven(parsed)
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
					_, _ = fmt.Fprintf(os.Stderr, "missing value for %s\n", s)
					os.Exit(1)
				}
				res[def.ArgName] = argsRead[i+1]
				i++ // skip the value
			} else {
				res[def.ArgName] = ""
			}
		} else {
			fmt.Printf("Unrecognized argument: %s\n", s)
		}
	}

	return res
}

func getArgs() []string {
	return os.Args[1:]
}

func defineArgs() []Arg {
	helpArg := Arg{
		ArgName:  "-h",
		ArgDesc:  "Show help message",
		FullArg:  "--help",
		HasValue: false,
	}

	configArg := Arg{
		ArgName:  "-c",
		ArgDesc:  "Specify path to config file (default: clay.yaml)",
		FullArg:  "--config",
		HasValue: true,
	}

	docsDirArg := Arg{
		ArgName:  "-d",
		ArgDesc:  "Specify path to documents directory (default: docs/)",
		FullArg:  "--docs-dir",
		HasValue: true,
	}

	outputArg := Arg{
		ArgName:  "-o",
		ArgDesc:  "Specify output directory (default: docs-out/)",
		FullArg:  "--output",
		HasValue: true,
	}

	forceArg := Arg{
		ArgName:  "-f",
		ArgDesc:  "Overwrite existing files",
		FullArg:  "--force",
		HasValue: false,
	}

	folderIconsArg := Arg{
		ArgName:  "-fm",
		ArgDesc:  "Specify path to folder meta file (default: dir-meta.yaml)",
		FullArg:  "--folder-meta",
		HasValue: true,
	}

	noConfirmArg := Arg{
		ArgName:  "-nc",
		ArgDesc:  "Do not ask for confirmation before overwriting files",
		FullArg:  "--no-confirm",
		HasValue: false,
	}

	args := []Arg{
		helpArg,
		configArg,
		docsDirArg,
		outputArg,
		forceArg,
		folderIconsArg,
		noConfirmArg,
	}

	return args
}
