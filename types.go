package ghaction

import "ghaction/go-gha-wrapper/ghapi"

type Manager string

const (
	ComposerManager Manager = "Composer"
)

type Action struct {
	client  *ghapi.Client
	manager Manager
	paths   pathInputs
	refs    refInputs
	repos   repoInputs
	options actionOptions
}

type pathInputs struct {
	lock string
	req  string
}
type refInputs struct {
	head string
	base string
}
type repoInputs struct {
	head string
	base string
}
type actionOptions struct {
	omitUnchanged   bool
	withStepSummary bool
	withAnnotations bool
}
