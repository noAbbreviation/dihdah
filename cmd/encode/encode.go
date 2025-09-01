package encode

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "encode",
	Short: "Drills for encoding morse code",
}

func init() {
	Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	Cmd.AddCommand(LetterCmd)
}
