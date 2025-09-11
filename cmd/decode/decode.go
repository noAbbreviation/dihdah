package decode

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "decode",
	Short: "Drills for decoding the morse code alphabet",
	Long: `This is the subcommand for decoding the morse code alphabet, from letters to sentences.

These are the three things the user can do in here:
  - 'dihdah decode letters': Gives the user drills to decode the morse code alphabet.
  - 'dihdah decode words': Gives the user drills to be proficient on decoding morse code words.
  - 'dihdah decode quotes': Gives the user drills to be proficient on decoding morse code sentences.

Run either 'dihdah decode letters --help', 'dihdah decode words --help', or
'dihdah decode quotes --help' for more details.`,
}

func init() {
	Cmd.AddCommand(LetterCmd)
	Cmd.AddCommand(WordCmd)
	Cmd.AddCommand(QuoteCmd)
}
