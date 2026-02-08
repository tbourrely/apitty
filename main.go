package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tbourrel/apitty/internal/model"
	"github.com/tbourrel/apitty/internal/ui"
)

func main() {
	m := model.InitialModel()

	// Set up Update and View from ui package
	p := tea.NewProgram(
		&appModel{m: m},
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// appModel wraps the model to provide Update and View methods
type appModel struct {
	m model.Model
}

func (a *appModel) Init() tea.Cmd {
	return a.m.Init()
}

func (a *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := ui.Update(a.m, msg)
	a.m = newModel
	return a, cmd
}

func (a *appModel) View() string {
	return ui.View(a.m)
}
