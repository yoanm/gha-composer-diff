package ghadepsdiff

type PkgManagerInput struct {
	// Lock represents the path to the lock file (e.g., composer.lock for composer, package-lock.json for npm,
	// yarn.lock for yarn, etc...)
	Lock string
	// Requirement represents the path to the requirement file  (e.g. composer.json for composer,
	// package.json for npm/yarn, etc...).
	Requirement string
}

type Config struct {
	Manager  ManagerType
	Previous PkgManagerInput
	Current  PkgManagerInput
}

type ManagerType string

const (
	ComposerManager ManagerType = "Composer"
)
