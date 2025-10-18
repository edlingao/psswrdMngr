package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/edlingao/psswrdMngr/configurator"
)

func main() {
	config := configurator.NewConfig().
		AddPassword().
		AddGroup().
		AddSecured().
		Start()

	initialModel := config.GetInitialModel()

	if _, err := tea.NewProgram(initialModel).Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
