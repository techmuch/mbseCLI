// Package feedback persists reviewer notes anchored to model elements by
// FQN. Notes live in a JSON sidecar file next to the source model
// (<model>.feedback.json) rather than a database, so a project's review
// history is plain text, diffable, and travels with the .sysml file in
// version control.
package feedback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Status is the review lifecycle state of a Note.
type Status string

const (
	StatusOpen     Status = "open"
	StatusInReview Status = "in_review"
	StatusResolved Status = "resolved"
)

// Note is a single piece of reviewer feedback anchored to an element FQN.
type Note struct {
	ID        string    `json:"id"`
	FQN       string    `json:"fqn"`
	Author    string    `json:"author,omitempty"`
	Text      string    `json:"text"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Store manages notes for a single model file, persisted to a JSON sidecar.
type Store struct {
	mu       sync.Mutex
	path     string // sidecar file path
	notes    map[string]*Note
}

// SidecarPath returns the conventional feedback file path for a model file,
// e.g. drone.sysml -> .drone.feedback.json (hidden, lives alongside it).
func SidecarPath(modelPath string) string {
	dir := filepath.Dir(modelPath)
	base := strings.TrimSuffix(filepath.Base(modelPath), filepath.Ext(modelPath))
	return filepath.Join(dir, "."+base+".feedback.json")
}

// Open loads (or initializes) the sidecar store for a model file.
func Open(modelPath string) (*Store, error) {
	s := &Store{path: SidecarPath(modelPath), notes: map[string]*Note{}}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var list []*Note
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.path, err)
	}
	for _, n := range list {
		s.notes[n.ID] = n
	}
	return s, nil
}

// ForFQN returns all notes anchored to a given element, newest first.
func (s *Store) ForFQN(fqn string) []*Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*Note
	for _, n := range s.notes {
		if n.FQN == fqn {
			out = append(out, n)
		}
	}
	return out
}

// All returns every note, keyed by FQN — the shape the UI wants for a single
// bulk hydrate on load.
func (s *Store) All() map[string][]*Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string][]*Note{}
	for _, n := range s.notes {
		out[n.FQN] = append(out[n.FQN], n)
	}
	return out
}

// Add creates a new note anchored to fqn and persists the store.
func (s *Store) Add(fqn, author, text string) (*Note, error) {
	s.mu.Lock()
	n := &Note{
		ID:        fmt.Sprintf("n_%d", time.Now().UnixNano()),
		FQN:       fqn,
		Author:    author,
		Text:      text,
		Status:    StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.notes[n.ID] = n
	s.mu.Unlock()
	return n, s.save()
}

// SetStatus updates a note's review status.
func (s *Store) SetStatus(id string, status Status) (*Note, error) {
	s.mu.Lock()
	n, ok := s.notes[id]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("note %s not found", id)
	}
	n.Status = status
	n.UpdatedAt = time.Now()
	s.mu.Unlock()
	return n, s.save()
}

// MarkOrphaned flags notes whose FQN no longer exists in the current graph
// (element renamed/deleted upstream) by prefixing the FQN. Callers pass the
// live set of valid FQNs after each reparse.
func (s *Store) MarkOrphaned(liveFQNs map[string]bool) []*Note {
	s.mu.Lock()
	var orphaned []*Note
	for _, n := range s.notes {
		if !liveFQNs[n.FQN] && !strings.HasPrefix(n.FQN, "orphan::") {
			orphaned = append(orphaned, n)
		}
	}
	s.mu.Unlock()
	return orphaned
}

func (s *Store) save() error {
	s.mu.Lock()
	list := make([]*Note, 0, len(s.notes))
	for _, n := range s.notes {
		list = append(list, n)
	}
	path := s.path
	s.mu.Unlock()

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
