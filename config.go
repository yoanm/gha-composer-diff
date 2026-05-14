package action

type Config struct {
	inputs *Inputs
	env    *Env
}

func NewConfig(inputs *Inputs, env *Env) *Config {
	return &Config{
		inputs: inputs,
		env:    env,
	}
}

type Inputs struct {
	lockPath        string
	reqPath         string
	prevRef         string
	currRef         string
	headRepo        string
	omitUnchanged   bool
	withStepSummary bool
	withAnnotations bool
	ghToken         string // Keep this field private to avoid accidental logging !!
}

func NewInputsConfig(
	lockPath, reqPath string,
	prevRef, currRef string,
	headRepo string,
	omitUnchanged bool,
	withStepSummary bool,
	withAnnotations bool,
	ghToken string,
) *Inputs {
	return &Inputs{
		lockPath:        lockPath,
		reqPath:         reqPath,
		prevRef:         prevRef,
		currRef:         currRef,
		headRepo:        headRepo,
		omitUnchanged:   omitUnchanged,
		withStepSummary: withStepSummary,
		withAnnotations: withAnnotations,
		ghToken:         ghToken,
	}
}

type Env struct {
	ghAPIUrl     string
	ghRepository string
}

func NewEnvConfig(ghAPIUrl, ghRepository string) *Env {
	return &Env{
		ghAPIUrl:     ghAPIUrl,
		ghRepository: ghRepository,
	}
}
