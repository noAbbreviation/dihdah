package commons

type TrainingModel struct {
	Drills       []Drill
	CurrentDrill int

	Correct []bool
}

type Drill struct {
	Text string

	Correct []bool
	Current int
}
