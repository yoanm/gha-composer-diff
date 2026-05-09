package main

type config struct {
	inputs *actionInputs
	env    *actionEnv
}

type actionInputs struct {
	lockPath        string
	reqPath         string
	prevRef         string
	currRef         string
	withStepSummary bool
	ghToken         string // Keep this field as private to avoid accidental logging !!
}

type actionEnv struct {
	ghRepository string
	ghAPIUrl     string
}
