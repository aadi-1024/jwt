package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	minLines = 1
	maxLines = 4
)

type App struct {
	inputs []textarea.Model
	//0 for JWT, 1 for Header, 2 for Claims, 3 for Secret Key
	editFlag int
	output   string
}

func initApp() App {
	app := App{}
	app.inputs = make([]textarea.Model, 4)

	for i := 0; i < 4; i++ {
		var _, placeholder string
		t := textarea.New()

		switch i {
		case 0:
			// prompt = "JWT     => "
			placeholder = "Token"
		case 1:
			// prompt = "Header  => "
			placeholder = "Header"
		case 2:
			// prompt = "Claims  => "
			placeholder = "Claims"
		case 3:
			// prompt = "Secrets => "
			placeholder = "Secrets"
		}

		// app.inputs[i].Prompt = prompt
		t.Placeholder = placeholder
		app.inputs[i] = t
	}

	app.inputs[0].Focus()
	return app
}

func (a App) changeFocus(old, new int) {
	a.inputs[old].Blur()
	a.inputs[new].Focus()
}

func (a App) updateContent() {
	if a.editFlag == 0 {
		spl := strings.Split(a.inputs[0].Value(), ".")
		if len(spl) != 3 {
			fmt.Println("here")
			for i := 1; i < 4; i++ {
				a.inputs[i].Reset()
			}
			return
		}

		header, err := base64.RawURLEncoding.DecodeString(spl[0])
		if err != nil {
			fmt.Println(err.Error())
			for i := 1; i < 4; i++ {
				a.inputs[i].Reset()
			}
			return
		}

		claims, err := base64.RawURLEncoding.DecodeString(spl[1])
		if err != nil {
			fmt.Println(err.Error())
			for i := 1; i < 4; i++ {
				a.inputs[i].Reset()
			}
			return
		}

		a.inputs[1].SetValue(string(header))
		a.inputs[2].SetValue(string(claims))
		return
	}
}

func (a App) Init() tea.Cmd {
	return textarea.Blink
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return a, tea.Quit

		case tea.KeyCtrlLeft:
			newFlag := max(0, a.editFlag-1)
			a.changeFocus(a.editFlag, newFlag)
			a.editFlag = newFlag
			return a, textarea.Blink

		case tea.KeyCtrlRight:
			newflag := min(3, a.editFlag+1)
			a.changeFocus(a.editFlag, newflag)
			a.editFlag = newflag
			return a, textarea.Blink

		case tea.KeyEnter:
			a.updateContent()
			return a, textarea.Blink
		}

	}

	var cmd tea.Cmd
	a.inputs[a.editFlag], cmd = a.inputs[a.editFlag].Update(msg)
	return a, cmd
}

func (a App) View() string {
	return fmt.Sprintf(
		"\n\n%s\n\n%s\n\n%s\n\n%s",
		a.inputs[0].View(),
		a.inputs[1].View(),
		a.inputs[2].View(),
		a.inputs[3].View(),
	) + "\n"
}
