package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/edlingao/psswrdMngr/configurator"
	"github.com/edlingao/psswrdMngr/internal/config"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatal("Failed to initialize config: ", err)
	}

	cfg := configurator.NewConfig().
		AddPassword().
		AddGroup().
		AddSecured().
		Start()

	initialModel := cfg.GetInitialModel()

	if _, err := tea.NewProgram(initialModel).Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
