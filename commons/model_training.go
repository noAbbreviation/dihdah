package commons

type TrainingModel struct {
	drills       []Drill
	currentDrill int

	correct []bool
}

type Drill struct {
	Text string

	Correct []bool
	Current int
}
