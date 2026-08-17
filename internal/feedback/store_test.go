package feedback

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSidecarPath(t *testing.T) {
	got := SidecarPath("/models/drone.sysml")
	want := filepath.Join("/models", ".drone.feedback.json")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func newTempModel(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "drone.sysml")
	if err := os.WriteFile(modelPath, []byte("package Drone {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	return modelPath
}

func TestStore_OpenOnMissingSidecarIsEmpty(t *testing.T) {
	s, err := Open(newTempModel(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.All()) != 0 {
		t.Errorf("expected empty store, got %d entries", len(s.All()))
	}
}

func TestStore_AddPersistsAndReloads(t *testing.T) {
	modelPath := newTempModel(t)

	s, err := Open(modelPath)
	if err != nil {
		t.Fatal(err)
	}
	note, err := s.Add("Drone::Airframe", "david", "check this")
	if err != nil {
		t.Fatal(err)
	}
	if note.Status != StatusOpen {
		t.Errorf("status = %q, want %q", note.Status, StatusOpen)
	}

	// A fresh Store reading the same sidecar should see the persisted note —
	// this is the property that makes review history survive a restart.
	reloaded, err := Open(modelPath)
	if err != nil {
		t.Fatal(err)
	}
	notes := reloaded.ForFQN("Drone::Airframe")
	if len(notes) != 1 {
		t.Fatalf("expected 1 note after reload, got %d", len(notes))
	}
	if notes[0].Text != "check this" {
		t.Errorf("text = %q, want %q", notes[0].Text, "check this")
	}
}

func TestStore_SetStatus(t *testing.T) {
	s, err := Open(newTempModel(t))
	if err != nil {
		t.Fatal(err)
	}
	note, err := s.Add("X::Y", "", "note")
	if err != nil {
		t.Fatal(err)
	}

	updated, err := s.SetStatus(note.ID, StatusResolved)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != StatusResolved {
		t.Errorf("status = %q, want %q", updated.Status, StatusResolved)
	}

	if _, err := s.SetStatus("does-not-exist", StatusResolved); err == nil {
		t.Error("expected an error for an unknown note id")
	}
}

func TestStore_MarkOrphaned(t *testing.T) {
	s, err := Open(newTempModel(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("Drone::Airframe", "", "still here"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("Drone::OldPart", "", "renamed away"); err != nil {
		t.Fatal(err)
	}

	live := map[string]bool{"Drone::Airframe": true}
	orphaned := s.MarkOrphaned(live)
	if len(orphaned) != 1 {
		t.Fatalf("expected 1 orphan, got %d", len(orphaned))
	}
	if orphaned[0].FQN != "Drone::OldPart" {
		t.Errorf("fqn = %q, want Drone::OldPart", orphaned[0].FQN)
	}
}

func TestStore_All(t *testing.T) {
	s, err := Open(newTempModel(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("A", "", "n1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("A", "", "n2"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("B", "", "n3"); err != nil {
		t.Fatal(err)
	}

	all := s.All()
	if len(all["A"]) != 2 {
		t.Errorf("A notes = %d, want 2", len(all["A"]))
	}
	if len(all["B"]) != 1 {
		t.Errorf("B notes = %d, want 1", len(all["B"]))
	}
}
