package wrapper

type Config struct {
	inputs *ActionInputs
	env    *ActionEnv
}

func NewConfig(inputs *ActionInputs, env *ActionEnv) *Config {
	return &Config{
		inputs: inputs,
		env:    env,
	}
}

type ActionInputs struct {
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

func NewActionInputs(
	lockPath, reqPath string,
	prevRef, currRef string,
	headRepo string,
	omitUnchanged bool,
	withStepSummary bool,
	withAnnotations bool,
	ghToken string,
) *ActionInputs {
	return &ActionInputs{
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

type ActionEnv struct {
	ghAPIUrl     string
	ghRepository string
}

func NewActionEnv(ghAPIUrl, ghRepository string) *ActionEnv {
	return &ActionEnv{
		ghAPIUrl:     ghAPIUrl,
		ghRepository: ghRepository,
	}
}
