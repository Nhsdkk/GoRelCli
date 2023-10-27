package main

import (
	"GoRelCli/generate"
	"GoRelCli/migrate"
	"flag"
	"fmt"
	"os"
)

func listAvailableCommands() {
	fmt.Println("List of available subcommands")
	fmt.Println("\t- migrate (runs migrations with options provided in gorel_schema.yml)")
	fmt.Println("\t- generate (generates go structs with options provided in gorel_schema.yml)")
}

func getFlags(args []string) (path string, output string) {
	fs := flag.FlagSet{}
	pathPtr := fs.String("path", "", "Path to GoRelCli schema file")
	outputPtr := fs.String("project_path", "", "Path to folder where generated files will be located")
	if err := fs.Parse(args); err != nil {
		fmt.Println("Error while parsing flags")
		os.Exit(3)
	}
	return *pathPtr, *outputPtr
}

func getHandler(args []string) func() error {
	if len(args) == 0 {
		fmt.Println("No subcommand provided")
		listAvailableCommands()
		os.Exit(3)
	}

	if args[0][0:1] == "-" {
		fmt.Println("No subcommand provided")
		listAvailableCommands()
		os.Exit(3)
	}

	path, output := getFlags(args[1:])

	switch args[0] {
	case "migrate":
		return func() error {
			return migrate.Migrate(path, output)
		}
	case "generate":
		return func() error {
			return generate.Generate(path, output)
		}
	default:
		fmt.Println(fmt.Sprintf("Command with name '%s' not found", args[0]))
		listAvailableCommands()
		os.Exit(3)
	}
	return func() error { return nil }
}

func main() {
	handler := getHandler(os.Args[1:])
	if err := handler(); err != nil {
		os.Exit(3)
	}

}
