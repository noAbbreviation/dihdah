package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "ui",
	Short: "Opens the TUI for this application",
	RunE: func(cmd *cobra.Command, args []string) error {
		model := newDihdahModel()
		p := tea.NewProgram(model, tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("Error running the model: %v", err)
		}
		return nil
	},
}

const RootCmdLong string = `This command line application focuses on providing drills to the user to be proficient on
decoding most of the International Morse Code characters (specifically only
the latin alphabet parts--the letters a to z).

    NOTE: Command line too intimidating? Run 'dihdah ui'.

First off, two caveats:
  - Encode drills are only for learning the morse code alphabet, not for learning
    the timings of how to send a morse code signal. The terminal does not give a consistent
    interface for detecting how long a keypress is held, so this is an unfortunate situation :(
  - For the convenience of the user, the application uses comma{,} as the dashes and the period{.} as the dot.

These are the two main things the user can do:
  - 'dihdah encode': Gives the user drills to learn how to write the morse code alphabet (the letters a-z).
  - 'dihdah decode': Gives the user drills to be proficient in interpreting morse code sounds.

Run either 'dihdah help encode' or 'dihdah help decode' for more details.
The user can also run 'dihdah ui' for a more user-friendly interface.`
