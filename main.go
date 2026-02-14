package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("tkz " + version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	if err := os.MkdirAll(getConfigDir(), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config directory: %v\n", err)
		os.Exit(1)
	}

	bwSession := os.Getenv("BW_SESSION")
	os.Unsetenv("BW_SESSION")

	p := tea.NewProgram(initialModel(bwSession), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("tkz - OAuth Token Manager")
	fmt.Println()
	fmt.Println("Manage OAuth clients and retrieve bearer tokens for development.")
	fmt.Println("Secrets are fetched from Bitwarden at runtime, never stored locally.")
	fmt.Println()
	fmt.Println("Usage: tkz [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println("  --version, -v    Show version")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  BW_SESSION       Bitwarden session key (optional, tkz prompts if needed)")
	fmt.Println()
	fmt.Println("Key Bindings:")
	fmt.Println("  enter            Get token for selected client")
	fmt.Println("  a                Add new OAuth client")
	fmt.Println("  e                Edit selected client")
	fmt.Println("  d                Delete selected client")
	fmt.Println("  /                Filter clients")
	fmt.Println("  q                Quit")
}
