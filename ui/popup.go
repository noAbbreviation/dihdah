package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Popup struct {
	message       []string
	backReference tea.Model
}

func (p Popup) Init() tea.Cmd {
	return nil
}

func (p Popup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, isKey := msg.(tea.KeyMsg); isKey {
		if msg.String() == "ctrl+c" {
			return p, tea.Quit
		}

		return p.backReference, nil
	}

	return p, nil
}

func (p Popup) View() string {
	joinedMessage := strings.Join(
		p.message[:],
		"\n",
	)

	return "\n" + joinedMessage + "\n" + "\n" + "(popup) (press any key to go back)" + "\n"
}
