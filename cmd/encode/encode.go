package encode

import (
	"fmt"

	"github.com/noAbbreviation/dihdah/cmd"
	"github.com/spf13/cobra"
)

var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Drills for encoding morse code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello from sub-command")
	},
}

func init() {
	cmd.Cmd.AddCommand(encodeCmd)
	encodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
