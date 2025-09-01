package commons

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
