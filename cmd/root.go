package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dihdah",
	Short: "Drills for learning morse code characters",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Hello from dihdah")
		return nil
	},
}

func Execute() {
	err := Cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type Drill struct {
	text string

	correct []bool
	current int
}

type TrainingModel struct {
	drills       []Drill
	currentDrill int

	correct []bool
}
