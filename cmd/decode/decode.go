package decode

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "decode",
	Short: "Drills for decoding morse code",
}

func init() {
	Cmd.AddCommand(LetterCmd)
	Cmd.AddCommand(WordCmd)
	Cmd.AddCommand(QuoteCmd)
}
