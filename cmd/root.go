package cmd

import (
	"os"

	"github.com/noAbbreviation/dihdah/cmd/decode"
	"github.com/noAbbreviation/dihdah/cmd/encode"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dihdah",
	Short: "Drills for learning morse code characters",
	Long: `
This command line application focuses on providing drills to the user to be proficient on
decoding most of the International Morse Code characters (specifically only
the latin alphabet parts--the letters a to z).

    NOTE: Command line too intimidating? Run 'dihdah ui'.

First off, two caveats:
  - Encode drills are only for learning the morse code alphabet, not for learning
    the timings of how to send a morse code signal. The terminal does not give a consistent
    interface for detecting how long a keypress is held, so this is an unfortunate situation :(
  - For convenience of the user, the application uses comma{,} as the dashes and the period{.} as the dot.

These are the two main things the user can do:
  - 'dihdah encode': Gives the user drills to learn how to write the morse code alphabet (the letters a-z).
  - 'dihdah decode': Gives the user drills to be proficient in interpreting morse code sounds.

Run either 'dihdah help encode' or 'dihdah help decode' for more details.
The user can also run 'dihdah ui' for a more user-friendly interface.`,
	// TODO: Create a user interface for general users.
	// TODO: Embed quotes.txt and words.txt
}

func Execute() {
	err := Cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	Cmd.CompletionOptions.DisableDefaultCmd = true

	Cmd.AddCommand(encode.Cmd)
	Cmd.AddCommand(decode.Cmd)
}
