package cmd

import (
	"os"

	"github.com/noAbbreviation/dihdah/cmd/encode"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dihdah",
	Short: "Drills for learning morse code characters",
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
}
