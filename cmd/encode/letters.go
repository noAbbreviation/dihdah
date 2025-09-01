package encode

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
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
	m    *commons.TrainingModel
	done <-chan bool
}

func newLetterModel() letterModel {
	done := make(chan bool)
	go func() {
		message := "dooby"
		delayBuffer := commons.SoundAssets[commons.ShortDelay]

		firstMorseCode := commons.MorseCodeLookup[rune(message[0])]
		speaker.PlayAndWait(commons.MorseCharSound(firstMorseCode))

		for _, c := range message[1:] {
			morseCode := commons.MorseCodeLookup[c]
			speaker.PlayAndWait(beep.Seq(
				beep.Loop(3, delayBuffer.Streamer(0, delayBuffer.Len())),
				commons.MorseCharSound(morseCode),
			))
		}

		done <- true
	}()

	return letterModel{done: done}
}

func (_m letterModel) Init() tea.Cmd {
	return nil
}

func (_m letterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	<-_m.done
	return _m, tea.Quit
}

func (_m letterModel) View() string {
	return "(playing the thing test)\n"
}
