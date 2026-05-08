package ghadepsdiff

import compdiff "github.com/yoanm/go-composer-diff"

type Config struct {
	Previous *compdiff.FileInput
	Current  *compdiff.FileInput
}
