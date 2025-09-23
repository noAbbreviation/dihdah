package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/noAbbreviation/dihdah/cmd/decode"
	"github.com/noAbbreviation/dihdah/cmd/encode"
	"github.com/noAbbreviation/dihdah/components"
)

const (
	viewPortWidth  = 120
	viewPortHeight = 10
)

type dihdahModel struct {
	inputs []components.Input

	helpViewPort viewport.Model
	selected     int

	helpText *string

	currentScreen screenEnum

	encodeFields       [5]inputField
	decodeLetterFields [6]inputField
	decodeWordFields   [5]inputField
	decodeQuoteFields  [2]inputField
}

type screenEnum int

const (
	mainScreen screenEnum = iota
	mainHelp

	encodeOptScreen
	encodeHelp

	decodeScreen
	decodeHelp

	decodeLetterOptScreen
	decodeLHelp

	decodeWordOptScreen
	decodeWHelp

	decodeQuoteOptScreen
	decodeQHelp
)

type mainScreenOpts int

const (
	encodeSelectM mainScreenOpts = iota
	decodeSelectM
	helpSelectM
	quitSelectM
)

type decodeScreenOpts int

const (
	decodeLetterSelectD decodeScreenOpts = iota
	decodeWordSelectD
	decodeQuoteSelectD

	decodeHelpSelectD
	backSelectD
)

type inputsE int

const (
	recapIE inputsE = iota
	customIE

	letterLevelIE
	wordLevelIE
	speedIE
	iterationsIE
	maxWordLengthIE

	lettersIE

	fileNameIE
)

func (input inputsE) String() string {
	return [...]string{
		"recap",
		"custom",
		"letterLevel",
		"wordLevel",
		"speed",
		"iterations",
		"maxWordLength",
		"letters",
		"fileName",
	}[input]
}

const (
	_ int = iota

	startButtonOffset
	helpButtonOffset
	backButtonOffset
)

type encodeIE int

const (
	encode__level_IE encodeIE = iota
	encode__interations_IE
	encode__recap_IE
	encode__custom_IE
	encode__letters_IE

	encode__start
	encode__help
	encode__back
)

func (inputEnum encodeIE) toInputEnum() inputsE {
	return [...]inputsE{
		letterLevelIE,
		iterationsIE,
		recapIE,
		customIE,
		lettersIE,
	}[inputEnum]
}

type decodeLettersIE int

const (
	decodeLetters__level_IE decodeLettersIE = iota
	decodeLetters__iterations_IE
	decodeLetters__recap_IE
	decodeLetters__custom_IE
	decodeLetters__letters_IE
	decodeLetters__speed_IE

	decodeLetters__start
	decodeLetters__help
	decodeLetters__back
)

func (inputEnum decodeLettersIE) toInputEnum() inputsE {
	return [...]inputsE{
		letterLevelIE,
		iterationsIE,
		recapIE,
		customIE,
		lettersIE,
		speedIE,
	}[inputEnum]
}

type decodeWordsIE int

const (
	decodeWords__custom_IE decodeWordsIE = iota
	decodeWords__maxLen_IE
	decodeWords__level_IE
	decodeWords__wordFile_IE
	decodeWords__speed_IE

	decodeWords__start
	decodeWords__help
	decodeWords__back
)

func (inputEnum decodeWordsIE) toInputEnum() inputsE {
	return [...]inputsE{
		customIE,
		maxWordLengthIE,
		wordLevelIE,
		fileNameIE,
		speedIE,
	}[inputEnum]
}

type decodeQuotesIE int

const (
	decodeQuotes__speed_IE decodeQuotesIE = iota
	decodeQuotes__quoteFile_IE

	decodeQuotes__start
	decodeQuotes__help
	decodeQuotes__back
)

func (inputEnum decodeQuotesIE) toInputEnum() inputsE {
	return [...]inputsE{
		speedIE,
		fileNameIE,
	}[inputEnum]
}

func viewPortInitContent(viewport *viewport.Model, text *string) {
	viewport.SetContent(lipgloss.NewStyle().Width(viewport.Width).Render(*text))
	viewport.GotoTop()
}

func validateLetters(s string) error {
	for _, r := range s {
		if r <= 'a' && r >= 'z' {
			return nil
		}

		if r <= 'A' && r >= 'Z' {
			return nil
		}
	}

	return fmt.Errorf("Essentially has no input")
}

func newDihdahModel() *dihdahModel {
	viewport := viewport.New(viewPortWidth, viewPortHeight)
	emptyStr := ""

	encodeFields := [...]inputField{
		{Prefix: "Level"},        // Show: !customChecked
		{Prefix: "  Iterations"}, // Show: !recapChecked
		{Prefix: "Recap?"},
		{Prefix: "Custom letters?"},
		{Prefix: "  Letters to use"}, // Show: customChecked
	}

	for i := range encode__back - encodeIE(backButtonOffset) + 1 {
		encodeFields[i].ReuseIndex = i
		encodeFields[i].Show = true
	}

	decodeLetterFields := [...]inputField{
		{Prefix: "Level"},        // Show: !customChecked
		{Prefix: "  Iterations"}, // Show: !recapChecked
		{Prefix: "Recap?"},
		{Prefix: "Custom letters?"},
		{Prefix: "  Letters to use"}, // Show: customChecked
		{Prefix: "Speed", Show: true},
	}

	for i := range decodeLetters__back - decodeLettersIE(backButtonOffset) + 1 {
		decodeLetterFields[i].ReuseIndex = i
		decodeLetterFields[i].Show = true
	}

	decodeWordFields := [...]inputField{
		{Prefix: "Custom word length?"},
		{Prefix: "  Level"},           // Show: !customChecked
		{Prefix: "  Max word length"}, // Show: customChecked
		{Prefix: "Custom word file"},
		{Prefix: "Speed"},
	}

	for i := range decodeWords__back - decodeWordsIE(backButtonOffset) + 1 {
		decodeWordFields[i].ReuseIndex = i
		decodeWordFields[i].Show = true
	}

	decodeQuoteFields := [...]inputField{
		{Prefix: "Speed", Show: true},
		{Prefix: "Custom quote file", Show: true},
	}

	for i := range decodeQuotes__back - decodeQuotesIE(backButtonOffset) + 1 {
		decodeQuoteFields[i].ReuseIndex = i
		decodeQuoteFields[i].Show = true
	}

	return &dihdahModel{
		inputs:       initInputs(),
		helpViewPort: viewport,
		helpText:     &emptyStr,

		encodeFields:       encodeFields,
		decodeLetterFields: decodeLetterFields,
		decodeWordFields:   decodeWordFields,
		decodeQuoteFields:  decodeQuoteFields,
	}
}

// TODO: Update iteration function for better granularity

func initInputs() []components.Input {
	inputs := make([]components.Input, fileNameIE+1)

	inputs[recapIE] = components.NewCheckBox(false)
	inputs[customIE] = components.NewCheckBox(false)
	inputs[letterLevelIE] = components.NewNumber(1, 7)
	inputs[wordLevelIE] = components.NewNumber(1, 4)

	inputs[speedIE] = components.NewNumber(0.25, 3)
	inputs[speedIE].(*components.Number).Default = 1
	inputs[speedIE].(*components.Number).SetDelta(0.25)

	inputs[iterationsIE] = components.NewNumber(1, 1<<16)
	inputs[iterationsIE].(*components.Number).Default = 3

	inputs[maxWordLengthIE] = components.NewNumber(3, 32)

	{
		textInput := textinput.New()
		textInput.CharLimit = 32
		textInput.Width = 16
		textInput.Validate = validateLetters
		textInput.Prompt = ""
		textInput.Placeholder = "the fox"

		inputs[lettersIE] = &components.TextInput{
			Input:            textInput,
			InvalidHighlight: "?",
		}
	}

	inputs[fileNameIE] = components.NewFilePicker(16, ".txt")
	return inputs
}

func (_m *dihdahModel) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(_m.inputs))
	for _, input := range _m.inputs {
		cmds = append(cmds, input.Init())
	}

	return tea.Sequence(textinput.Blink, tea.Batch(cmds...))
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

	uiNavigate := false
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		_m.helpViewPort.Width = min(msg.Width, viewPortWidth)
		// NOTE: The height's "magic number" refers at the extra helpView render at View(...)
		_m.helpViewPort.Height = min(msg.Height-5, viewPortHeight)

		_m.helpViewPort.SetContent(lipgloss.NewStyle().Width(_m.helpViewPort.Width).Render(*_m.helpText))
	case tea.KeyMsg:
		// TODO: Special keymsg cases would want to fallthrough to input handling pipeline
		switch msg.String() {
		case "esc":
			if _m.currentScreen == mainScreen {
				return _m, tea.Quit
			}

			switch _m.currentScreen {
			case mainHelp:
				_m.currentScreen = mainScreen

			case encodeOptScreen:
				_m.currentScreen = mainScreen
			case encodeHelp:
				_m.currentScreen = encodeOptScreen

			case decodeScreen:
				_m.currentScreen = mainScreen
			case decodeHelp:
				_m.currentScreen = decodeScreen

			case decodeLHelp:
				_m.currentScreen = decodeLetterOptScreen
			case decodeWHelp:
				_m.currentScreen = decodeWordOptScreen
			case decodeQHelp:
				_m.currentScreen = decodeQuoteOptScreen

			case decodeLetterOptScreen:
				_m.currentScreen = decodeScreen
			case decodeWordOptScreen:
				_m.currentScreen = decodeScreen
			case decodeQuoteOptScreen:
				_m.currentScreen = decodeScreen
			}

			isUiScreen, _ := _m.uiMaxIndex(_m.currentScreen)
			if !isUiScreen {
				_m.selected = 0
				break
			}

			inputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
			if !inputScreen {
				_m.selected = 0
				break
			}

			indexes := _m.renderedInputIndexes(_m.currentScreen)
			_m.selected = indexes[0]

			uiNavigate = true
		case "enter":
			isUiScreen, _ := _m.uiMaxIndex(_m.currentScreen)
			if !isUiScreen {
				break
			}

			switch _m.currentScreen {
			case encodeOptScreen:
				indexes := _m.renderedInputIndexes(_m.currentScreen)
				lastIdx := indexes[len(indexes)-1]

				switch _m.selected {
				case lastIdx + backButtonOffset:
					_m.currentScreen = mainScreen

				case lastIdx + helpButtonOffset:
					helpText := encode.Cmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = encodeHelp

				case lastIdx + startButtonOffset:
					// TODO: glue here
				}

			case decodeLetterOptScreen:
				indexes := _m.renderedInputIndexes(_m.currentScreen)
				lastIdx := indexes[len(indexes)-1]

				switch _m.selected {
				case lastIdx + backButtonOffset:
					_m.currentScreen = decodeScreen

				case lastIdx + helpButtonOffset:
					helpText := decode.LetterCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeLHelp

				case lastIdx + startButtonOffset:
					// TODO: glue here
				}

			case decodeWordOptScreen:
				indexes := _m.renderedInputIndexes(_m.currentScreen)
				lastIdx := indexes[len(indexes)-1]

				switch _m.selected {
				case lastIdx + backButtonOffset:
					_m.currentScreen = decodeScreen

				case lastIdx + helpButtonOffset:
					helpText := decode.WordCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeWHelp

				case lastIdx + startButtonOffset:
					// TODO: glue here
				}

			case decodeQuoteOptScreen:
				indexes := _m.renderedInputIndexes(_m.currentScreen)
				lastIdx := indexes[len(indexes)-1]

				switch _m.selected {
				case lastIdx + backButtonOffset:
					_m.currentScreen = decodeScreen

				case lastIdx + helpButtonOffset:
					helpText := decode.QuoteCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeQHelp

				case lastIdx + startButtonOffset:
					// TODO: glue here
				}

			case mainScreen:
				switch mainScreenOpts(_m.selected) {
				case encodeSelectM:
					_m.currentScreen = encodeOptScreen
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

				default:
					fallthrough
				case backSelectD:
					_m.currentScreen = mainScreen
				}
			}

			inputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
			if !inputScreen {
				_m.selected = 0
				break
			}

			indexes := _m.renderedInputIndexes(_m.currentScreen)
			_m.selected = indexes[0]

			uiNavigate = true
		case "down", "ctrl+n", "shift+tab":
			isUiScreen, _ := _m.uiMaxIndex(_m.currentScreen)

			if !isUiScreen {
				break
			}

			_m.navigateDown()
			uiNavigate = true
		case "up", "ctrl+p", "tab":
			isUiScreen, _ := _m.uiMaxIndex(_m.currentScreen)
			if !isUiScreen {
				break
			}

			_m.navigateUp()
			uiNavigate = true
		default:
			uiScreen, _ := _m.uiMaxIndex(_m.currentScreen)
			if !uiScreen {
				break
			}

			inputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
			inputI := _m.toInputIE(_m.currentScreen, _m.selected)

			exoticNavigation := true

			switch {
			default:
				exoticNavigation = false

			case uiScreen && !inputScreen && msg.String() == "j":
				fallthrough
			case inputScreen && (inputI != lettersIE) && msg.String() == "j":
				_m.navigateDown()

			case uiScreen && !inputScreen && msg.String() == "k":
				fallthrough
			case inputScreen && (inputI != lettersIE) && msg.String() == "k":
				_m.navigateUp()
			}

			if exoticNavigation {
				uiNavigate = true
				break
			}

			if !inputScreen {
				break
			}

			inputs := _m.renderedInputIndexes(_m.currentScreen)
			if _m.selected > inputs[len(inputs)-1] {
				return _m, tea.Batch(cmds...)
			}

			focusedInputE := _m.toInputIE(_m.currentScreen, _m.selected)
			doDefaultUpdate := true

			switch input := _m.inputs[focusedInputE].(type) {
			case *components.Number:
				switch msg.String() {
				case "+", "=", ".", ">", "right", "l":
					input.Increment()
					doDefaultUpdate = false

				case "-", "_", ",", "<", "left", "h":
					input.Decrement()
					doDefaultUpdate = false
				}
			}

			switch msg.String() {
			case "ctrl+r":
				_m.inputs[focusedInputE].Reset()
				doDefaultUpdate = false
			}

			if doDefaultUpdate {
				var cmd tea.Cmd

				_m.inputs[focusedInputE], cmd = _m.inputs[focusedInputE].Update(msg)
				cmds = append(cmds, cmd)
			}

			_m.updateInputUI()
		}
	default:
		inputScreen, inputMaxIdx := inputsRawMaxIdx(_m.currentScreen)
		if !inputScreen {
			break
		}

		if _m.selected > inputMaxIdx {
			break
		}

		var cmd tea.Cmd

		focusedInputE := _m.toInputIE(_m.currentScreen, _m.selected)
		_m.inputs[focusedInputE], cmd = _m.inputs[focusedInputE].Update(msg)
		_m.updateInputUI()

		cmds = append(cmds, cmd)
		return _m, tea.Batch(cmds...)
	}

	inputScreen, inputMaxIdx := inputsRawMaxIdx(_m.currentScreen)
	if !inputScreen {
		return _m, tea.Batch(cmds...)
	}

	if uiNavigate {
		for i := range inputMaxIdx + 1 {
			inputE := _m.toInputIE(_m.currentScreen, i)
			_m.inputs[inputE].Blur()
		}

		inputs := _m.renderedInputIndexes(_m.currentScreen)
		if _m.selected <= inputs[len(inputs)-1] {
			focusedInputE := _m.toInputIE(_m.currentScreen, _m.selected)

			cmd := _m.inputs[focusedInputE].Focus()
			cmds = append(cmds, cmd)
		}

		_m.updateInputUI()
	}

	return _m, tea.Batch(cmds...)
}

func (_m *dihdahModel) navigateUp() {
	uiScreen, maxIdx := _m.uiMaxIndex(_m.currentScreen)
	if !uiScreen {
		return
	}

	oldSelected := _m.selected

	_m.selected -= 1
	if _m.selected < 0 {
		_m.selected = maxIdx
	}

	inputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
	if !inputScreen {
		return
	}

	indexes := _m.renderedInputIndexes(_m.currentScreen)
	if _m.selected < indexes[len(indexes)-1] {
		reverseIndex := 0
		for i, inputE := range slices.Backward(indexes) {
			if oldSelected == inputE {
				reverseIndex = i
				break
			}
		}

		_m.selected = indexes[reverseIndex-1]
	}
}

func (_m *dihdahModel) navigateDown() {
	uiScreen, maxIdx := _m.uiMaxIndex(_m.currentScreen)
	if !uiScreen {
		return
	}

	oldSelected := _m.selected
	wrappedAround := false

	_m.selected += 1
	if _m.selected > maxIdx {
		wrappedAround = true
		_m.selected = 0
	}

	inputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
	if !inputScreen {
		return
	}

	indexes := _m.renderedInputIndexes(_m.currentScreen)
	if wrappedAround {
		_m.selected = indexes[0]
		return
	}

	if _m.selected < indexes[len(indexes)-1] {
		reverseIndex := 0
		for i, inputE := range indexes {
			if oldSelected == inputE {
				reverseIndex = i
				break
			}
		}

		_m.selected = indexes[reverseIndex+1]
		return
	}

	_m.selected = max(_m.selected, indexes[len(indexes)-1])
}

func (_m *dihdahModel) updateInputUI() {
	inputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
	if !inputScreen {
		return
	}

	// TODO: (and NOTE that) Level field can change inputs
	switch _m.currentScreen {
	case encodeOptScreen:
		customChecked := _m.inputs[encode__custom_IE.toInputEnum()].Value().(bool)
		recapChecked := _m.inputs[encode__recap_IE.toInputEnum()].Value().(bool)

		_m.encodeFields[encode__level_IE].Show = !customChecked
		_m.encodeFields[encode__interations_IE].Show = !recapChecked
		_m.encodeFields[encode__letters_IE].Show = customChecked

	case decodeLetterOptScreen:
		customChecked := _m.inputs[decodeLetters__custom_IE.toInputEnum()].Value().(bool)
		recapChecked := _m.inputs[decodeLetters__recap_IE.toInputEnum()].Value().(bool)

		_m.decodeLetterFields[decodeLetters__level_IE].Show = !customChecked
		_m.decodeLetterFields[decodeLetters__iterations_IE].Show = !recapChecked
		_m.decodeLetterFields[decodeLetters__letters_IE].Show = customChecked

	case decodeWordOptScreen:
		customChecked := _m.inputs[decodeWords__custom_IE.toInputEnum()].Value().(bool)

		_m.decodeWordFields[decodeWords__level_IE].Show = !customChecked
		_m.decodeWordFields[decodeWords__maxLen_IE].Show = customChecked

	case decodeQuoteOptScreen:
		// empty implementation
	}
}

func (_m *dihdahModel) renderedInputIndexes(currentScreen screenEnum) []int {
	inputScreen, _ := inputsRawMaxIdx(currentScreen)
	if !inputScreen {
		return nil
	}

	inputFields := []inputField(nil)

	switch _m.currentScreen {
	case encodeOptScreen:
		inputFields = _m.encodeFields[:]

	case decodeLetterOptScreen:
		inputFields = _m.decodeLetterFields[:]

	case decodeWordOptScreen:
		inputFields = _m.decodeWordFields[:]

	case decodeQuoteOptScreen:
		inputFields = _m.decodeQuoteFields[:]
	}

	indexes := []int(nil)
	for i, input := range inputFields {
		if input.Show {
			indexes = append(indexes, i)
		}
	}

	return indexes
}

func (_m *dihdahModel) uiMaxIndex(currentScreen screenEnum) (bool, int) {
	isUiScreen := true
	maxIdx := 0

	switch currentScreen {
	default:
		isUiScreen = false
	case mainScreen:
		maxIdx = int(quitSelectM)
	case decodeScreen:
		maxIdx = int(backSelectD)

	case encodeOptScreen:
		fallthrough
	case decodeLetterOptScreen:
		fallthrough
	case decodeWordOptScreen:
		fallthrough
	case decodeQuoteOptScreen:
		inputs := _m.renderedInputIndexes(currentScreen)
		maxIdx = inputs[len(inputs)-1] + backButtonOffset
	}

	return isUiScreen, maxIdx
}

func inputsRawMaxIdx(currentScreen screenEnum) (bool, int) {
	maxIdx := 0
	isInputScreen := true

	switch currentScreen {
	default:
		isInputScreen = false
	case encodeOptScreen:
		maxIdx = int(encode__back) - backButtonOffset
	case decodeLetterOptScreen:
		maxIdx = int(decodeLetters__back) - backButtonOffset
	case decodeWordOptScreen:
		maxIdx = int(decodeWords__back) - backButtonOffset
	case decodeQuoteOptScreen:
		maxIdx = int(decodeQuotes__back) - backButtonOffset
	}

	return isInputScreen, maxIdx
}

func (_m dihdahModel) toInputIE(currentScreen screenEnum, localInputE int) inputsE {
	inputE := 0
	inputs := _m.renderedInputIndexes(currentScreen)
	if len(inputs) == 0 || localInputE > inputs[len(inputs)-1] {
		return 0
	}

	switch currentScreen {
	case encodeOptScreen:
		inputE = int(encodeIE(localInputE).toInputEnum())
	case decodeLetterOptScreen:
		inputE = int(decodeLettersIE(localInputE).toInputEnum())
	case decodeWordOptScreen:
		inputE = int(decodeWordsIE(localInputE).toInputEnum())
	case decodeQuoteOptScreen:
		inputE = int(decodeQuotesIE(localInputE).toInputEnum())
	}

	return inputsE(inputE)
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

	return strings.Join(renderedOptions, "\n")
}

type inputField struct {
	Show       bool
	Prefix     string
	ReuseIndex inputReuser
}

type inputReuser interface {
	toInputEnum() inputsE
}

func renderInputs(realInputs []components.Input, fields []inputField) string {
	displayedFields := []string{}

	for _, field := range fields {
		if !field.Show {
			continue
		}

		inputView := realInputs[field.ReuseIndex.toInputEnum()].View()
		render := fmt.Sprintf("  %v: %v", field.Prefix, inputView)
		displayedFields = append(displayedFields, render)
	}

	return strings.Join(displayedFields, "\n")
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

	optScreenHeader := ""
	inputFields := []inputField(nil)

	switch _m.currentScreen {
	case encodeOptScreen:
		inputFields = _m.encodeFields[:]
		optScreenHeader = "Encode letter training"

	case decodeLetterOptScreen:
		inputFields = _m.decodeLetterFields[:]
		optScreenHeader = "Decode letter training"

	case decodeWordOptScreen:
		inputFields = _m.decodeWordFields[:]
		optScreenHeader = "Decode word training"

	case decodeQuoteOptScreen:
		inputFields = _m.decodeQuoteFields[:]
		optScreenHeader = "Decode word training"
	}

	if len(inputFields) != 0 {
		indexes := _m.renderedInputIndexes(_m.currentScreen)
		offsettedSelected := _m.selected - (indexes[len(indexes)-1] + 1)
		renderedCommonOpts := renderOpts([]string{
			"Start training",
			"Help",
			"Back",
		}, offsettedSelected)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			optScreenHeader,
			"",
			renderInputs(_m.inputs, inputFields),
			"",
			renderedCommonOpts,
			"",
			"(training options) (up/down to navigate, left/right to change numbers, space to toggle, ctrl+r to reset selection)",
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

	case decodeScreen:
		renderedOptions = renderOpts([]string{
			"Start letter decode training",
			"Start word decode training",
			"Start quote decode training",
			"Help page for decode training",
			"Back to main menu",
		}, _m.selected)
		screenHeader = "Dihdah: Decode training"
	}

	isUiScreen, _ := _m.uiMaxIndex(_m.currentScreen)
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
