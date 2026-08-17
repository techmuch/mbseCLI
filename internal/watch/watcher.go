// Package watch monitors a directory for .sysml file changes and emits
// debounced change events. Debouncing absorbs editor save patterns (e.g.
// write-to-tmp-then-rename, multiple flushes) so the parser isn't invoked on
// a half-written file mid-save.
package watch

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Event describes a single settled change to a watched .sysml file.
type Event struct {
	Path string
	Op   string // "write", "create", "remove", "rename"
}

// Watcher watches a directory tree for .sysml changes and delivers debounced
// events on Events.
type Watcher struct {
	Events chan Event

	fsw      *fsnotify.Watcher
	debounce time.Duration
}

// New creates a Watcher rooted at dir (recursively) with the given debounce
// window (recommended: 150-300ms — long enough to skip partial writes,
// short enough to still feel "live").
func New(dir string, debounce time.Duration) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{Events: make(chan Event, 16), fsw: fsw, debounce: debounce}
	if err := addRecursive(fsw, dir); err != nil {
		fsw.Close()
		return nil, err
	}
	go w.loop()
	return w, nil
}

// addRecursive registers watches on dir and every subdirectory. fsnotify has
// no recursive-watch primitive, so new subdirectories created after startup
// are picked up in loop() when a Create event for a directory arrives.
func addRecursive(fsw *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return fsw.Add(path)
		}
		return nil
	})
}

func isSysML(path string) bool {
	return strings.HasSuffix(path, ".sysml")
}

// loop consumes raw fsnotify events, filters to .sysml files, and debounces
// bursts (multiple writes to the same path within the debounce window
// collapse into a single Event) before forwarding to Events.
func (w *Watcher) loop() {
	pending := make(map[string]*time.Timer)
	fire := make(chan Event, 16)

	for {
		select {
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			if ev.Has(fsnotify.Create) {
				if info, err := statIsDir(ev.Name); err == nil && info {
					_ = addRecursive(w.fsw, ev.Name)
				}
			}
			if !isSysML(ev.Name) {
				continue
			}
			op := opName(ev)
			path := ev.Name
			if t, exists := pending[path]; exists {
				t.Stop()
			}
			pending[path] = time.AfterFunc(w.debounce, func() {
				fire <- Event{Path: path, Op: op}
			})
		case ev := <-fire:
			delete(pending, ev.Path)
			w.Events <- ev
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			_ = err // surfaced via logging by the caller if desired
		}
	}
}

func opName(ev fsnotify.Event) string {
	switch {
	case ev.Has(fsnotify.Remove):
		return "remove"
	case ev.Has(fsnotify.Rename):
		return "rename"
	case ev.Has(fsnotify.Create):
		return "create"
	default:
		return "write"
	}
}

func statIsDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

// Close stops the underlying fsnotify watcher and closes Events.
func (w *Watcher) Close() error {
	err := w.fsw.Close()
	close(w.Events)
	return err
}
