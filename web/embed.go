// Package web embeds the built React frontend (web/dist, produced by
// `npm run build` in the web/ directory) into the mbsecli binary.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Assets returns the embedded frontend build rooted at dist/, ready to be
// served as a filesystem root. Returns an error if the frontend hasn't been
// built yet (dist/ only has the placeholder file checked into git).
func Assets() (fs.FS, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, err
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, err
	}
	return sub, nil
}
