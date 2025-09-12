package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/noAbbreviation/dihdah/cmd/decode"
	"github.com/noAbbreviation/dihdah/cmd/encode"
)

const (
	viewPortWidth  = 120
	viewPortHeight = 10
)

type dihdahModel struct {
	// TODO: input reuse
	inputs       []textinput.Model
	helpViewPort viewport.Model
	selected     int

	helpText *string

	currentScreen screenEnum
	mainScreenO   []mainScreenOpts
	encodeScreenO []encodeScreenOpts
	decodeScreenO []decodeScreenOpts
}

type screenEnum int

const (
	mainScreen screenEnum = iota
	mainHelp

	encodeScreen
	encodeHelp

	decodeScreen
	decodeOptScreen
	decodeHelp

	decodeLHelp
	decodeWHelp
	decodeQHelp

	encodeOptScreen
	decodeLetterOptScreen
	decodeWordOptScreen
	decodeQuoteOptScreen
)

type mainScreenOpts int

const (
	encodeSelectM mainScreenOpts = iota
	decodeSelectM
	helpSelectM
	quitSelectM
)

type encodeScreenOpts int

const (
	encodeSelectE encodeScreenOpts = iota
	helpSelectE
	backSelectE
)

type decodeScreenOpts int

const (
	decodeLetterSelectD decodeScreenOpts = iota
	decodeWordSelectD
	decodeQuoteSelectD

	decodeHelpSelectD
	decodeLetterHelpSelectD
	decodeWordHelpSelectD
	decodeQuoteHelpSelectD

	backSelectD
)

func viewPortInitContent(viewport *viewport.Model, text *string) {
	viewport.SetContent(lipgloss.NewStyle().Width(viewport.Width).Render(*text))
	viewport.GotoTop()
}

func newDihdahModel() *dihdahModel {
	viewport := viewport.New(viewPortWidth, viewPortHeight)

	mainScreenO := make([]mainScreenOpts, quitSelectM+1)
	for i := range quitSelectM {
		mainScreenO[i] = i
	}

	encodeScreenO := make([]encodeScreenOpts, backSelectE+1)
	for i := range backSelectE {
		encodeScreenO[i] = i
	}

	decodeScreenO := make([]decodeScreenOpts, backSelectD+1)
	for i := range backSelectD {
		decodeScreenO[i] = i
	}

	emptyStr := ""
	return &dihdahModel{
		helpViewPort:  viewport,
		helpText:      &emptyStr,
		mainScreenO:   mainScreenO,
		encodeScreenO: encodeScreenO,
		decodeScreenO: decodeScreenO,
	}
}

func (_m *dihdahModel) Init() tea.Cmd {
	return textinput.Blink
}

func (_m *dihdahModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, isKey := msg.(tea.KeyMsg); isKey && key.Type == tea.KeyCtrlC {
		return _m, tea.Quit
	}

	cmds := []tea.Cmd(nil)
	isHelp := false

	switch _m.currentScreen {
	case decodeHelp, decodeLHelp, decodeWHelp, decodeQHelp,
		mainHelp, encodeHelp:
		isHelp = true
	}

	if isHelp {
		var cmd tea.Cmd
		_m.helpViewPort, cmd = _m.helpViewPort.Update(msg)

		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		_m.helpViewPort.Width = min(msg.Width, viewPortWidth)
		// NOTE: The height's "magic number" refers at the extra helpView render at View(...)
		_m.helpViewPort.Height = min(msg.Height-5, viewPortHeight)

		_m.helpViewPort.SetContent(lipgloss.NewStyle().Width(_m.helpViewPort.Width).Render(*_m.helpText))
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace", "esc":
			if _m.currentScreen == mainScreen {
				return _m, tea.Quit
			}

			switch _m.currentScreen {
			case mainHelp:
				_m.currentScreen = mainScreen
			case encodeScreen:
				_m.currentScreen = mainScreen
			case encodeHelp:
				_m.currentScreen = encodeScreen

			case decodeScreen:
				_m.currentScreen = mainScreen
			case decodeHelp:
				_m.currentScreen = decodeScreen

			case decodeLHelp:
				_m.currentScreen = decodeScreen
			case decodeWHelp:
				_m.currentScreen = decodeScreen
			case decodeQHelp:
				_m.currentScreen = decodeScreen

			case encodeOptScreen:
				_m.currentScreen = encodeScreen
			case decodeLetterOptScreen:
				_m.currentScreen = decodeScreen
			case decodeWordOptScreen:
				_m.currentScreen = decodeScreen
			case decodeQuoteOptScreen:
				_m.currentScreen = decodeScreen
			}

			_m.selected = 0
			return _m, nil
		case "enter":
			isUiScreen, _ := uiMaxIndex(_m.currentScreen)

			if !isUiScreen {
				break
			}

			switch _m.currentScreen {
			default:
				fallthrough

			case mainScreen:
				switch mainScreenOpts(_m.selected) {
				case encodeSelectM:
					_m.currentScreen = encodeScreen
				case decodeSelectM:
					_m.currentScreen = decodeScreen
				case helpSelectM:
					helpText := RootCmdLong
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.helpText = &helpText
					_m.currentScreen = mainHelp
				default:
					fallthrough
				case quitSelectM:
					return _m, tea.Quit
				}

			case encodeScreen:
				switch encodeScreenOpts(_m.selected) {
				case encodeSelectE:
					_m.currentScreen = encodeOptScreen
				case helpSelectE:
					helpText := encode.Cmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = encodeHelp
				default:
					fallthrough
				case backSelectE:
					_m.currentScreen = mainScreen
				}

			case decodeScreen:
				switch decodeScreenOpts(_m.selected) {
				case decodeLetterSelectD:
					_m.currentScreen = decodeLetterOptScreen
				case decodeWordSelectD:
					_m.currentScreen = decodeWordOptScreen
				case decodeQuoteSelectD:
					_m.currentScreen = decodeQuoteOptScreen

				case decodeHelpSelectD:
					helpText := decode.Cmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeHelp

				case decodeLetterHelpSelectD:
					helpText := decode.LetterCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeHelp

				case decodeQuoteHelpSelectD:
					helpText := decode.WordCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeHelp

				case decodeWordHelpSelectD:
					helpText := decode.QuoteCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeHelp

				default:
					fallthrough
				case backSelectD:
					_m.currentScreen = mainScreen
				}
			}

			_m.selected = 0
		case "down", "j", "ctrl+n", "shift+tab":
			isUiScreen, maxIdx := uiMaxIndex(_m.currentScreen)

			if !isUiScreen {
				break
			}

			_m.selected = (_m.selected + 1) % (maxIdx + 1)
		case "up", "k", "ctrl+p", "tab":
			isUiScreen, maxIdx := uiMaxIndex(_m.currentScreen)

			if !isUiScreen {
				break
			}

			_m.selected -= 1
			if _m.selected < 0 {
				_m.selected = maxIdx
			}
		}
	}

	return _m, tea.Batch(cmds...)
}

func uiMaxIndex(currentScreen screenEnum) (bool, int) {
	isUiScreen := true
	maxIdx := 0

	switch currentScreen {
	default:
		isUiScreen = false
	case mainScreen:
		maxIdx = int(quitSelectM)
	case encodeScreen:
		maxIdx = int(backSelectE)
	case decodeScreen:
		maxIdx = int(backSelectD)
	}

	return isUiScreen, maxIdx
}

func renderOpts(options []string, selected int) string {
	renderedOptions := []string{}

	for i, option := range options {
		selectedStr := " "
		if i == selected {
			selectedStr = "+"
		}

		renderedOpt := fmt.Sprintf("  [%v] %v", selectedStr, option)
		renderedOptions = append(renderedOptions, renderedOpt)
	}

	//  NOTE: Could optimize this for strings.Join(...)
	return lipgloss.JoinVertical(lipgloss.Left, renderedOptions...)
}

func (_m *dihdahModel) View() string {
	isHelp := true
	helpViewTopic := ""

	switch _m.currentScreen {
	case mainHelp:
		helpViewTopic = "Dihdah overview"
	case encodeHelp:
		helpViewTopic = "Encode Command"

	case decodeHelp:
		helpViewTopic = "Decode Command"
	case decodeLHelp:
		helpViewTopic = "Decode Letter Command"
	case decodeWHelp:
		helpViewTopic = "Decode Word Command"
	case decodeQHelp:
		helpViewTopic = "Decode Quote Command"

	default:
		isHelp = false
	}

	if isHelp {
		helpViewHeader := fmt.Sprintf("Help: %v", helpViewTopic)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			helpViewHeader,
			strings.Repeat("-", _m.helpViewPort.Width),
			_m.helpViewPort.View(),
			strings.Repeat("-", viewPortWidth),
			"(help page) (up/down to navigate, backspace/esc to go back, ctrl+c to exit)",
			"",
		)
	}

	var screenHeader string
	var renderedOptions string

	switch _m.currentScreen {
	default:
		fallthrough
	case mainScreen:
		renderedOptions = renderOpts([]string{
			"Encode training",
			"Decode training",
			"Help page",
			"Quit application",
		}, _m.selected)
		screenHeader = "Dihdah (github.com/noAbbreviation/dihdah)"

	case encodeScreen:
		renderedOptions = renderOpts([]string{
			"Start encode training",
			"Help page",
			"Back to main menu",
		}, _m.selected)
		screenHeader = "Dihdah: Encode training"

	case decodeScreen:
		renderedOptions = renderOpts([]string{
			"Start letter decode training",
			"Start word decode training",
			"Start quote decode training",
			"Help page for decode training",
			"Help page for letter decode training",
			"Help page for word decode training",
			"Help page for quote decode training",
			"Back to main menu",
		}, _m.selected)
		screenHeader = "Dihdah: Decode training"
	}

	isUiScreen, _ := uiMaxIndex(_m.currentScreen)
	if isUiScreen {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			screenHeader,
			"",
			renderedOptions,
			"",
			"(ui screen) (up/down to navigate, enter to select, backspace/esc to go back, ctrl+c to exit)",
			"",
		)
	}

	return "(dihdah ui) (under construction)"
}
