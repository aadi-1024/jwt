package main

import (
	"encoding/json"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/golang-jwt/jwt/v5"
)

var (
	textAreaStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#dddddd"))
	textAreaBlurred = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#666666"))
)

type App struct {
	inputs []textarea.Model
	//0 for JWT, 1 for Header, 2 for Claims, 3 for Secret Key
	editFlag int
	output   string
	error    bool
	width    int
	height   int
}

func initApp() App {
	app := App{}
	app.inputs = make([]textarea.Model, 4)
	app.error = false

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
		t.Prompt = ""
		t.Placeholder = placeholder
		t.FocusedStyle.Base = textAreaStyle
		t.BlurredStyle.Base = textAreaBlurred
		t.ShowLineNumbers = false
		t.Blur()
		app.inputs[i] = t
	}

	app.inputs[0].Focus()
	return app
}

func (a App) changeFocus(old, new int) {
	a.inputs[old].Blur()
	a.inputs[new].Focus()
}

func (a App) updateContent() (string, error) {
	var errr error
	errr = nil
	if a.editFlag == 0 {
		jwtToken := a.inputs[0].Value()
		m := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(jwtToken, m, func(t *jwt.Token) (interface{}, error) {
			return []byte(a.inputs[3].Value()), nil
		})

		if err != nil {
			errr = err
		}
		if token == nil {
			return "", errr
		}

		header, err := json.MarshalIndent(token.Header, "", "\t")
		if err != nil {
			errr = err
		}
		claims, err := json.MarshalIndent(m, "", "\t")
		if err != nil {
			errr = err
		}

		a.inputs[1].SetValue(string(header))
		a.inputs[2].SetValue(string(claims))
	} else {
		header := make(map[string]any)
		claims := make(map[string]any)

		if err := json.Unmarshal([]byte(a.inputs[1].Value()), &header); err != nil {
			errr = err
		}

		if err := json.Unmarshal([]byte(a.inputs[2].Value()), &claims); err != nil {
			errr = err
		}

		token := jwt.New(jwt.SigningMethodHS256)
		token.Header = header
		token.Claims = jwt.MapClaims(claims)

		str, err := token.SignedString([]byte(a.inputs[3].Value()))
		if err != nil {
			errr = err
			str = ""
		}

		a.inputs[0].SetValue(str)
	}

	return "Verified Successfully", errr
}

func (a App) Init() tea.Cmd {
	return textarea.Blink
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	a.error = false
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlQ:
			return a, tea.Quit

		case tea.KeyCtrlC:
			for i := range a.inputs {
				a.inputs[i].Reset()
			}
			return a, textarea.Blink

		case tea.KeyCtrlLeft, tea.KeyCtrlUp:
			newFlag := max(0, a.editFlag-1)
			a.changeFocus(a.editFlag, newFlag)
			a.editFlag = newFlag
			return a, textarea.Blink

		case tea.KeyCtrlRight, tea.KeyCtrlDown:
			newflag := min(3, a.editFlag+1)
			a.changeFocus(a.editFlag, newflag)
			a.editFlag = newflag
			return a, textarea.Blink

		case tea.KeyCtrlZ:
			_, err := a.updateContent()
			if err != nil {
				a.output = err.Error()
				a.error = true
			} else {
				a.output = "Verified"
			}
			return a, textarea.Blink

		case tea.KeyCtrlX:
			err := clipboard.WriteAll(a.inputs[0].Value())
			if err != nil {
				a.output = err.Error()
				a.error = true
			}
			return a, textarea.Blink
		}

	case tea.WindowSizeMsg:
		a.height = msg.Height
		a.width = msg.Width
	}

	var cmd tea.Cmd
	a.inputs[a.editFlag], cmd = a.inputs[a.editFlag].Update(msg)
	return a, cmd
}

func (a App) View() string {
	var color string
	if a.error {
		color = "#ffbbbb"
	} else {
		color = "#bbffbb"
	}
	vertical := lipgloss.NewStyle().Width(a.width / 3).Render(lipgloss.JoinVertical(lipgloss.Center, a.inputs[1].View(), a.inputs[2].View(), a.inputs[3].View()))
	return lipgloss.NewStyle().Height(a.height).AlignVertical(lipgloss.Center).Render(lipgloss.JoinHorizontal(lipgloss.Center, lipgloss.NewStyle().Width(a.width/3).Render(a.inputs[0].View()), vertical, lipgloss.JoinVertical(lipgloss.Center, lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Foreground(lipgloss.Color(color)).Width(a.width/3).Render(a.output), lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).AlignHorizontal(lipgloss.Center).Render("Keybinds:\n\nCtrl+Z: calculate, Ctrl+X: copy token\n\nCtrl+C: clear, Ctrl+Arrow: navigate\n\n Ctrl+Q: quit"))))
}
