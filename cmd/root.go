package cmd

import (
	"os"

	"github.com/noAbbreviation/dihdah/cmd/decode"
	"github.com/noAbbreviation/dihdah/cmd/encode"
	"github.com/noAbbreviation/dihdah/ui"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dihdah",
	Short: "Drills for learning morse code characters",
	Long:  ui.RootCmdLong,
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
	Cmd.AddCommand(ui.Cmd)
}
