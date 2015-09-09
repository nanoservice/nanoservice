//go:generate nanotemplate -T command.Command -t command -I github.com/nanoservice/nanoservice/command --input=_fetcher.tt.go
package main // import "github.com/nanoservice/nanoservice"

import (
	"fmt"
	"github.com/nanoservice/nanoservice/command"
	"github.com/nanoservice/nanoservice/configure"
	"github.com/nanoservice/nanoservice/create"
	"github.com/nanoservice/nanoservice/deploy"
	"github.com/nanoservice/nanoservice/fetcher_command"
	"github.com/nanoservice/nanoservice/scale"
	"os"
)

var (
	invalidCommand = command.Command(func([]string) {
		fmt.Fprintf(os.Stderr, "Unknown command '%s'\n\n", commandName())
		printAvailableCommands()
		os.Exit(1)
	})

	rawCommands = map[string]command.Command{
		"configure": configure.Command,
		"create":    create.Command,
		"deploy":    deploy.Command,
		"scale":     scale.Command,
	}

	commands = fetcher_command.
			New(&rawCommands).
			WithDefault(&invalidCommand)
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	fetchCommand()(commandArgs())
}

func commandName() string {
	return os.Args[1]
}

func commandArgs() []string {
	return os.Args[2:]
}

func fetchCommand() command.Command {
	return commands.Fetch(commandName())
}

func printUsage() {
	fmt.Printf("Usage: %s command [options]\n\n", os.Args[0])
	printAvailableCommands()
}

func printAvailableCommands() {
	fmt.Fprintf(os.Stderr, "Available commands:\n")
	for name, _ := range rawCommands {
		fmt.Fprintf(os.Stderr, "\t%s\n", name)
	}
}
