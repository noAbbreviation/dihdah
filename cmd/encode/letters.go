package encode

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/noAbbreviation/dihdah/commons"
	"github.com/spf13/cobra"
)

var LetterCmd = &cobra.Command{
	Use:   "letter",
	Short: "Train for decoding letters.",
	Run: func(_cmd *cobra.Command, args []string) {
		p := tea.NewProgram(newLetterModel())

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running the program: %v", err)
			os.Exit(1)
		}
	},
}

type letterModel struct {
	m *commons.TrainingModel
}

func newLetterModel() letterModel {
	return letterModel{}
}

type doneMsg struct{}

func playMorseCode() tea.Msg {
	speed := float64(1)
	delayBuffer := commons.SoundAssets[commons.ShortDelay]
	delayStreamer := delayBuffer.Streamer(0, delayBuffer.Len())
	delayResampler := beep.ResampleRatio(4, speed, delayStreamer)

	emptyStreamer := beep.NewBuffer(commons.AudioFormat)
	emptyStreamer.Append(delayResampler)

	message := "the quick brown fox jumps over the lazy dog"

	firstMorseCode := commons.MorseCodeLookup[rune(message[0])]
	speaker.PlayAndWait(commons.MorseCharSound(firstMorseCode, speed))

	for _, c := range message[1:] {
		streamer := emptyStreamer.Streamer(0, emptyStreamer.Len())

		if c == ' ' {
			speaker.PlayAndWait(beep.Loop(7, streamer))
			continue
		}

		morseCode := commons.MorseCodeLookup[c]

		speaker.PlayAndWait(beep.Loop(3, streamer))
		speaker.PlayAndWait(commons.MorseCharSound(morseCode, speed))
	}

	return doneMsg{}
}

func (_m letterModel) Init() tea.Cmd {
	return playMorseCode
}

func (_m letterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return _m, tea.Quit
		}

	case doneMsg:
		return _m, tea.Quit
	}

	return _m, nil
}

func (_m letterModel) View() string {
	return "(playing the thing test)\n"
}
