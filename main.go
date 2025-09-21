package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/noAbbreviation/dihdah/components"
)

// import "github.com/noAbbreviation/dihdah/cmd"

func main() {
	// cmd.Execute()

	p := tea.NewProgram(newTest(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type compTest struct {
	inputs   [4]components.Input
	selected inputE
}

type inputE int

const (
	numberPicker inputE = iota
	checkboxPicker
	filePicker
	textInputPicker

	checkValues
)

func (value inputE) String() string {
	return [...]string{
		"numberPicker",
		"checkboxPicker",
		"filePicker",
		"textInputPicker",
		"checkValues",
	}[value]
}

func newTest() *compTest {
	inputs := [4]components.Input{}
	{
		numberInput := components.NewNumber(1, 12)
		numberInput.SetDelta(2)

		inputs[numberPicker] = numberInput
	}

	inputs[checkboxPicker] = components.NewCheckBox(false)
	inputs[filePicker] = components.NewFilePicker(20, ".txt")

	{
		textInput := textinput.New()
		textInput.CharLimit = 30
		textInput.Width = 20
		textInput.Prompt = ""
		textInput.Placeholder = "the fox"

		inputs[textInputPicker] = &components.TextInput{Input: textInput}
	}

	return &compTest{
		inputs: inputs,
	}
}

func (m *compTest) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.inputs))
	for _, input := range m.inputs {
		cmds = append(cmds, input.Init())
	}

	m.inputs[0].Focus()
	return tea.Batch(
		textinput.Blink,
		tea.Batch(cmds...),
	)
}

func (m *compTest) NavigateUp() {
	m.selected -= 1
	if m.selected < 0 {
		m.selected = checkValues
	}
}

func (m *compTest) NavigateDown() {
	m.selected += 1
	if m.selected > checkValues {
		m.selected = 0
	}
}

func (m *compTest) selectingAFile() (*components.FilePicker, bool) {
	filePicker := m.inputs[filePicker].(*components.FilePicker)
	return filePicker, filePicker.SelectingFile
}

func (m *compTest) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd(nil)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	if _, selectingFile := m.selectingAFile(); selectingFile {
		var cmd tea.Cmd

		m.inputs[filePicker], cmd = m.inputs[filePicker].Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	default:
		if m.selected >= checkValues {
			break
		}

		var cmd tea.Cmd

		m.inputs[m.selected], cmd = m.inputs[m.selected].Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		keyMsg := msg.String()

		if m.selected == checkValues && keyMsg == "enter" {
			cmds = append(cmds, tea.Printf(
				"numberValue: %v, checkboxValue: %v, filePicker value: %v, textInput value: %v\n",
				m.inputs[numberPicker].Value(),
				m.inputs[checkboxPicker].Value(),
				m.inputs[filePicker].Value(),
				m.inputs[textInputPicker].Value(),
			))
			break
		}

		navigationMotion := true
		switch keyMsg {
		case "up", "shift+tab", "ctrl+p":
			m.NavigateUp()
		case "down", "tab", "ctrl+n":
			m.NavigateDown()
		default:
			_, isText := m.inputs[m.selected].(*components.TextInput)

			exoticNavigation := true
			switch {
			case !isText && keyMsg == "j":
				m.NavigateDown()
			case !isText && keyMsg == "k":
				m.NavigateUp()
			default:
				exoticNavigation = false
			}

			if !exoticNavigation {
				navigationMotion = false
			}
		}

		if navigationMotion {
			for _, input := range m.inputs {
				input.Blur()
			}

			if m.selected < checkValues {
				cmd := m.inputs[m.selected].Focus()
				cmds = append(cmds, cmd)
			}

			break
		}

		switch input := m.inputs[m.selected].(type) {
		case *components.Number:
			switch keyMsg {
			case "+", "=", ".", ">":
				input.Increment()
			case "-", "_", ",", "<":
				input.Decrement()
			}
		default:
			var cmd tea.Cmd

			m.inputs[m.selected], cmd = m.inputs[m.selected].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *compTest) View() string {
	if filePicker, selectingFile := m.selectingAFile(); selectingFile {
		return filePicker.View()
	}

	inputs := [len(m.inputs)]string{
		"Level: ",
		"Is thingy?: ",
		"Word file: ",
		"Message: ",
	}

	for i, headerStr := range inputs {
		inputs[i] = headerStr + m.inputs[i].View()
	}

	checkValueSelectedStr := " "
	if m.selected == checkValues {
		checkValueSelectedStr = "+"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Join(inputs[:], "\n"),
		fmt.Sprintf("check the thing here [%v]", checkValueSelectedStr),
	)
}
