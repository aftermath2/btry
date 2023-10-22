package ui

import (
	"embed"
	"io/fs"

	"github.com/pkg/errors"
)

// Dir contains the compiled frontend code.
//
//go:embed all:dist
var ui embed.FS

// FS returns a filesystem with the UI's files embedded.
func FS() (fs.FS, error) {
	u, err := fs.Sub(ui, "dist")
	if err != nil {
		return nil, errors.Wrap(err, "embedding files")
	}

	return u, nil
}
