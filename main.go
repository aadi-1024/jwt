package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initApp(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err.Error())
	}
}
