package ui

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/noAbbreviation/dihdah/assets"
	"github.com/noAbbreviation/dihdah/cmd/decode"
	"github.com/noAbbreviation/dihdah/cmd/encode"
	"github.com/noAbbreviation/dihdah/components"
)

const (
	viewPortWidth  = 120
	viewPortHeight = 10
)

type updater[T any] interface {
	Update(tea.Msg) (T, tea.Cmd)
}

func updateModel[T updater[T]](cmds *[]tea.Cmd, model *T, msg tea.Msg) {
	var cmd tea.Cmd

	*model, cmd = (*model).Update(msg)
	*cmds = append(*cmds, cmd)
}

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
	decodeWords__level_IE
	decodeWords__maxLen_IE
	decodeWords__wordFile_IE
	decodeWords__speed_IE

	decodeWords__start
	decodeWords__help
	decodeWords__back
)

func (inputEnum decodeWordsIE) toInputEnum() inputsE {
	return [...]inputsE{
		customIE,
		wordLevelIE,
		maxWordLengthIE,
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
	if len(s) < 3 {
		return fmt.Errorf("Essentially has no input")
	}

	if len(encode.DedupCleanLetters(s)) < 3 {
		return fmt.Errorf("Essentially has no input")
	}

	return nil
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

func (_m dihdahModel) letterLevelUpdate() {
	levelArg := int(_m.inputs[letterLevelIE].Value().(float64))
	lettersPerLevel := encode.NewLettersPerLevel

	levelArg = min(levelArg, len(lettersPerLevel))

	letters := ""
	for _, newLetters := range lettersPerLevel[:levelArg] {
		letters += newLetters
	}

	dedupedLetters := encode.DedupCleanLetters(letters)
	_m.inputs[lettersIE].SetValue(dedupedLetters)

	iterations := max(float64(len(dedupedLetters)/2), 3)
	_m.inputs[iterationsIE].SetValue(iterations)
}

func (_m dihdahModel) wordLevelUpdate() {
	levelArg := int(_m.inputs[wordLevelIE].Value().(float64))

	wordLengths := decode.MaxWordLenPerLevel
	levelArg = min(levelArg, len(wordLengths))

	wordLength := float64(wordLengths[levelArg-1])
	_m.inputs[maxWordLengthIE].SetValue(wordLength)
}

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

	inputs[maxWordLengthIE] = components.NewNumber(3, 1<<16)

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
		cmd := input.Init()
		cmds = append(cmds, cmd)
	}

	_m.letterLevelUpdate()
	_m.wordLevelUpdate()

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
		updateModel(&cmds, &_m.helpViewPort, msg)
	}

	insideInput, focusedIdx := _m.toInputIE(_m.currentScreen, _m.selected)
	if insideInput && focusedIdx == fileNameIE {
		filePicker := _m.inputs[fileNameIE].(*components.FilePicker)

		if filePicker.SelectingFile {
			updateModel(&cmds, &_m.inputs[fileNameIE], msg)
			return _m, tea.Batch(cmds...)
		}
	}

	uiNavigate := false
	doNoOP := false

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		_m.helpViewPort.Width = min(msg.Width, viewPortWidth)
		// NOTE: The height's "magic number" refers at the extra helpView render at View(...)
		_m.helpViewPort.Height = min(msg.Height-5, viewPortHeight)

		_m.helpViewPort.SetContent(lipgloss.NewStyle().Width(_m.helpViewPort.Width).Render(*_m.helpText))
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "h", "left", "H", "shift+left":
			if _m.currentScreen == mainScreen {
				return _m, tea.Quit
			}

			inputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
			if inputScreen {
				specialCase := false
				insideInput, focusedIE := _m.toInputIE(_m.currentScreen, _m.selected)

				if !insideInput {
					goto afterSpecialBackChecks
				}

				switch input := _m.inputs[focusedIE].(type) {
				case *components.TextInput:
					switch msg.String() {
					case "h", "H", "backspace":
						updateModel(&cmds, &_m.inputs[focusedIE], msg)
						specialCase = true
					}

				case *components.FilePicker:
					switch msg.String() {
					case "h", "left":
						updateModel(&cmds, &_m.inputs[focusedIE], msg)
						specialCase = true
					}

				case *components.Number:
					switch msg.String() {
					case "h", "left":
						input.Decrement()

						switch focusedIE {
						case letterLevelIE:
							_m.letterLevelUpdate()
						case wordLevelIE:
							_m.wordLevelUpdate()
						}

						specialCase = true
					}
				}

			afterSpecialBackChecks:
				if specialCase {
					break
				}
			}

			switch _m.currentScreen {
			default:
				doNoOP = true

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

			if doNoOP {
				break
			}

			freshInputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
			if !freshInputScreen {
				_m.selected = 0
				break
			}

			indexes := _m.renderedInputIndexes(_m.currentScreen)
			_m.selected = indexes[0]

			uiNavigate = true
		case "enter", " ", "l", "right", "L", "shift+right":
			isUiScreen, _ := _m.uiMaxIndex(_m.currentScreen)
			if !isUiScreen {
				break
			}

			inputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
			if inputScreen {
				specialCase := false
				insideInput, focusedIE := _m.toInputIE(_m.currentScreen, _m.selected)

				if !insideInput {
					goto afterSpecialForwardChecks
				}

				switch input := _m.inputs[focusedIE].(type) {
				case *components.Checkbox:
					switch msg.String() {
					case "enter", " ":
						input.Toggle()
						specialCase = true
					}

				case *components.TextInput:
					switch msg.String() {
					case "l", "L", " ":
						updateModel(&cmds, &_m.inputs[focusedIE], msg)
						specialCase = true
					}

				case *components.FilePicker:
					switch msg.String() {
					case "enter", " ", "l", "right":
						updateModel(&cmds, &_m.inputs[focusedIE], msg)
						specialCase = true
					}

				case *components.Number:
					switch msg.String() {
					case "l", "right":
						input.Increment()

						switch focusedIE {
						case letterLevelIE:
							_m.letterLevelUpdate()
						case wordLevelIE:
							_m.wordLevelUpdate()
						}

						specialCase = true
					}
				}

			afterSpecialForwardChecks:
				if specialCase {
					break
				}
			}

			switch _m.currentScreen {
			case encodeOptScreen:
				indexes := _m.renderedInputIndexes(_m.currentScreen)
				lastIdx := indexes[len(indexes)-1]

				switch _m.selected {
				default:
					doNoOP = true

				case lastIdx + backButtonOffset:
					_m.currentScreen = mainScreen

				case lastIdx + helpButtonOffset:
					helpText := encode.Cmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = encodeHelp

				case lastIdx + startButtonOffset:
					lettersInput := _m.inputs[lettersIE].(*components.TextInput).Input
					if lettersInput.Err != nil {
						doNoOP = true
						break
					}

					letters := lettersInput.Value()
					if len(letters) == 0 {
						doNoOP = true
						break
					}

					dedupedLetters := encode.DedupCleanLetters(letters)
					runes := []rune(dedupedLetters)

					trainingLetters := ""

					toRecap := _m.inputs[recapIE].Value().(bool)
					if toRecap {
						rand.Shuffle(len(runes), func(i, j int) {
							runes[i], runes[j] = runes[j], runes[i]
						})

						trainingLetters = string(runes)
						encodeModel := encode.NewLetterModel(trainingLetters, _m)
						return encodeModel, encodeModel.Init()
					}

					iterations := _m.inputs[iterationsIE].Value().(float64)
					for range int(iterations) {
						letter := runes[rand.Intn(len(runes))]
						trainingLetters += string(letter)
					}

					encodeModel := encode.NewLetterModel(trainingLetters, _m)
					return encodeModel, encodeModel.Init()
				}

			case decodeLetterOptScreen:
				indexes := _m.renderedInputIndexes(_m.currentScreen)
				lastIdx := indexes[len(indexes)-1]

				switch _m.selected {
				default:
					doNoOP = true

				case lastIdx + backButtonOffset:
					_m.currentScreen = decodeScreen

				case lastIdx + helpButtonOffset:
					helpText := decode.LetterCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeLHelp

				case lastIdx + startButtonOffset:
					lettersInput := _m.inputs[lettersIE].(*components.TextInput).Input
					if lettersInput.Err != nil {
						doNoOP = true
						break
					}

					letters := lettersInput.Value()
					dedupedLetters := encode.DedupCleanLetters(letters)
					runes := []rune(dedupedLetters)

					trainingLetters := ""
					speed := _m.inputs[speedIE].Value().(float64)

					toRecap := _m.inputs[recapIE].Value().(bool)
					if toRecap {
						rand.Shuffle(len(runes), func(i, j int) {
							runes[i], runes[j] = runes[j], runes[i]
						})

						trainingLetters = string(runes)
						decodeWordsM := decode.NewLetterModel(trainingLetters, dedupedLetters, speed, _m)
						return decodeWordsM, decodeWordsM.Init()
					}

					iterations := _m.inputs[iterationsIE].Value().(float64)
					for range int(iterations) {
						letter := runes[rand.Intn(len(runes))]
						trainingLetters += string(letter)
					}

					decodeWordsM := decode.NewLetterModel(trainingLetters, dedupedLetters, speed, _m)
					return decodeWordsM, decodeWordsM.Init()
				}

			case decodeWordOptScreen:
				indexes := _m.renderedInputIndexes(_m.currentScreen)
				lastIdx := indexes[len(indexes)-1]

				switch _m.selected {
				default:
					doNoOP = true

				case lastIdx + backButtonOffset:
					_m.currentScreen = decodeScreen

				case lastIdx + helpButtonOffset:
					helpText := decode.WordCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeWHelp

				case lastIdx + startButtonOffset:
					maxWordLen := _m.inputs[maxWordLengthIE].Value().(float64)
					wordFile := _m.inputs[fileNameIE].Value().(string)

					var wordsReader io.Reader

					if len(wordFile) == 0 {
						wordsReader = strings.NewReader(assets.Words)
					} else {
						file, err := os.Open(wordFile)
						if err != nil {
							return Popup{message: []string{
								"File cannot be found :(",
								fmt.Sprintf("File name: %v", wordFile),
							}, backReference: _m}, nil
						}

						defer file.Close()
						wordsReader = file
					}
					maxWordLens := decode.MaxWordLenPerLevel

					wordPool := []string(nil)
					scanner := bufio.NewScanner(wordsReader)

					scanner.Split(bufio.ScanWords)
					for scanner.Scan() {
						word := scanner.Text()
						word = strings.Map(func(r rune) rune {
							if r >= 'a' && r <= 'z' {
								return r
							}

							if r >= 'A' && r <= 'Z' {
								return r
							}

							if r == '-' {
								return r
							}

							return -1
						}, word)

						if len(word) <= int(maxWordLen) ||
							int(maxWordLen) >= maxWordLens[len(maxWordLens)-1] {

							wordPool = append(wordPool, word)
						}
					}

					// scanner.SplitWords(...) does not return errors!
					_ = scanner.Err()

					const iterations = 5
					words := make([]string, 0, iterations)

					if len(wordPool) < iterations {
						return Popup{message: []string{
							"Error creating the word training drill:",
							"There are less available words than training iterations :(",
						}, backReference: _m}, nil
					}

					for range iterations {
						wordIdx := rand.Intn(len(wordPool))
						words = append(words, wordPool[wordIdx])

						wordPool[wordIdx] = wordPool[len(wordPool)-1]
						wordPool = wordPool[:len(wordPool)-1]
					}

					speed := _m.inputs[speedIE].Value().(float64)
					decodeWModel := decode.NewWordModel(words[:], uint16(maxWordLen), speed, _m)

					return decodeWModel, decodeWModel.Init()
				}

			case decodeQuoteOptScreen:
				indexes := _m.renderedInputIndexes(_m.currentScreen)
				lastIdx := indexes[len(indexes)-1]

				switch _m.selected {
				default:
					doNoOP = true

				case lastIdx + backButtonOffset:
					_m.currentScreen = decodeScreen

				case lastIdx + helpButtonOffset:
					helpText := decode.QuoteCmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeQHelp

				case lastIdx + startButtonOffset:
					quoteFile := _m.inputs[fileNameIE].Value().(string)

					var quotesReader io.Reader

					if len(quoteFile) == 0 {
						quotesReader = strings.NewReader(assets.Words)
					} else {
						file, err := os.Open(quoteFile)
						if err != nil {
							return Popup{message: []string{
								"File cannot be found :(",
								fmt.Sprintf("File name: %v", quoteFile),
							}, backReference: _m}, nil
						}

						defer file.Close()
						quotesReader = file
					}

					quotes := []string(nil)
					scanner := bufio.NewScanner(quotesReader)

					scanner.Split(bufio.ScanLines)
					for scanner.Scan() {
						quote := strings.TrimSpace(scanner.Text())
						quotes = append(quotes, quote)
					}

					// scanner.SplitLines(...) does not return errors!
					_ = scanner.Err()

					if len(quotes) == 0 {
						return Popup{message: []string{
							"Error processing the quote file:",
							"The quote file is empty :(",
						}, backReference: _m}, nil
					}

					speed := _m.inputs[speedIE].Value().(float64)
					randomQuote := quotes[rand.Intn(len(quotes))]

					decodeQModel := decode.NewQuoteModel(randomQuote, speed, _m)
					return decodeQModel, decodeQModel.Init()
				}
				lipgloss.NewStyle().Render()

			case mainScreen:
				switch mainScreenOpts(_m.selected) {
				default:
					doNoOP = true

				case encodeSelectM:
					_m.currentScreen = encodeOptScreen
				case decodeSelectM:
					_m.currentScreen = decodeScreen
				case helpSelectM:
					helpText := RootCmdLong
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.helpText = &helpText
					_m.currentScreen = mainHelp
				case quitSelectM:
					return _m, tea.Quit
				}

			case decodeScreen:
				switch decodeScreenOpts(_m.selected) {
				default:
					doNoOP = true

				case decodeLetterSelectD:
					_m.currentScreen = decodeLetterOptScreen

				case decodeWordSelectD:
					_m.currentScreen = decodeWordOptScreen
					_m.inputs[fileNameIE].Reset()

					cmds = append(cmds, _m.inputs[fileNameIE].Init())

				case decodeQuoteSelectD:
					_m.currentScreen = decodeQuoteOptScreen
					_m.inputs[fileNameIE].Reset()

					cmds = append(cmds, _m.inputs[fileNameIE].Init())

				case decodeHelpSelectD:
					helpText := decode.Cmd.Long
					_m.helpText = &helpText
					viewPortInitContent(&_m.helpViewPort, &helpText)
					_m.currentScreen = decodeHelp

				case backSelectD:
					_m.currentScreen = mainScreen
				}
			}

			if doNoOP {
				break
			}

			freshInputScreen, _ := inputsRawMaxIdx(_m.currentScreen)
			if !freshInputScreen {
				_m.selected = 0
				break
			}

			indexes := _m.renderedInputIndexes(_m.currentScreen)
			_m.selected = indexes[0]

			uiNavigate = true
		case "down", "ctrl+n", "shift+tab", "j":
			isUiScreen, _ := _m.uiMaxIndex(_m.currentScreen)

			if !isUiScreen {
				break
			}

			specialCase := false
			insideInput, focusedIE := _m.toInputIE(_m.currentScreen, _m.selected)

			if insideInput {
				switch _m.inputs[focusedIE].(type) {
				case *components.TextInput:
					switch msg.String() {
					case "j":
						updateModel(&cmds, &_m.inputs[focusedIE], msg)
						specialCase = true
					}
				}
			}

			if specialCase {
				break
			}

			_m.navigateDown()
			uiNavigate = true
		case "up", "ctrl+p", "tab", "k":
			isUiScreen, _ := _m.uiMaxIndex(_m.currentScreen)

			if !isUiScreen {
				break
			}

			specialCase := false
			insideInput, focusedIE := _m.toInputIE(_m.currentScreen, _m.selected)

			if insideInput {
				switch _m.inputs[focusedIE].(type) {
				case *components.TextInput:
					switch msg.String() {
					case "k":
						updateModel(&cmds, &_m.inputs[focusedIE], msg)
						specialCase = true
					}
				}
			}

			if specialCase {
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
			if !inputScreen {
				break
			}

			inputs := _m.renderedInputIndexes(_m.currentScreen)
			if _m.selected > inputs[len(inputs)-1] {
				return _m, tea.Batch(cmds...)
			}

			insideInput, focusedInputE := _m.toInputIE(_m.currentScreen, _m.selected)
			if !insideInput {
				break
			}

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
				updateModel(&cmds, &_m.inputs[focusedInputE], msg)
			}

			switch focusedInputE {
			case letterLevelIE:
				_m.letterLevelUpdate()

			case wordLevelIE:
				_m.wordLevelUpdate()
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

		insideInput, focusedInputE := _m.toInputIE(_m.currentScreen, _m.selected)
		if insideInput {
			updateModel(&cmds, &_m.inputs[focusedInputE], msg)
		}

		_m.updateInputUI()
		return _m, tea.Batch(cmds...)
	}

	if doNoOP {
		return _m, tea.Batch(cmds...)
	}

	inputScreen, inputMaxIdx := inputsRawMaxIdx(_m.currentScreen)
	if !inputScreen {
		return _m, tea.Batch(cmds...)
	}

	defer _m.updateInputUI()
	if uiNavigate {
		for i := range inputMaxIdx + 1 {
			_, inputE := _m.toInputIE(_m.currentScreen, i)
			_m.inputs[inputE].Blur()
		}

		insideInput, focusedInputE := _m.toInputIE(_m.currentScreen, _m.selected)
		if insideInput {
			cmd := _m.inputs[focusedInputE].Focus()
			cmds = append(cmds, cmd)
		}
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

func (_m *dihdahModel) uiMaxIndex(currentScreen screenEnum) (isUiScreen bool, maxIdx int) {
	maxIdx = 0
	isUiScreen = true

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

func inputsRawMaxIdx(currentScreen screenEnum) (isInputScreen bool, maxIdx int) {
	maxIdx = 0
	isInputScreen = true

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

func (_m dihdahModel) toInputIE(currentScreen screenEnum, localInputE int) (insideInput bool, inputE inputsE) {
	inputE = 0
	inputs := _m.renderedInputIndexes(currentScreen)
	if len(inputs) == 0 || localInputE > inputs[len(inputs)-1] {
		return false, inputE
	}

	switch currentScreen {
	case encodeOptScreen:
		inputE = encodeIE(localInputE).toInputEnum()
	case decodeLetterOptScreen:
		inputE = decodeLettersIE(localInputE).toInputEnum()
	case decodeWordOptScreen:
		inputE = decodeWordsIE(localInputE).toInputEnum()
	case decodeQuoteOptScreen:
		inputE = decodeQuotesIE(localInputE).toInputEnum()
	}

	return true, inputE
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

	insideInput, focusedIdx := _m.toInputIE(_m.currentScreen, _m.selected)

	if insideInput && focusedIdx == fileNameIE {
		filePicker := _m.inputs[fileNameIE].(*components.FilePicker)
		if filePicker.SelectingFile {
			return filePicker.View()
		}
	}

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
			"(training options) (up/down to navigate, left/right to change numbers, space to toggle, ctrl+r to reset current input)",
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
