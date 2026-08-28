package docs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestClearPending(t *testing.T) {
	root := t.TempDir()
	if err := RecordPending(root, errors.New("boom")); err != nil {
		t.Fatal(err)
	}
	rec, err := LoadPending(root)
	if err != nil {
		t.Fatal(err)
	}
	if rec == nil || rec.Error != "boom" {
		t.Fatalf("expected recorded pending, got %+v", rec)
	}
	if err := clearPending(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".rpg", "pending.json")); !os.IsNotExist(err) {
		t.Fatalf("pending.json should be removed, stat err: %v", err)
	}
	// Idempotent: clearing again when absent is not an error.
	if err := clearPending(root); err != nil {
		t.Fatalf("clearPending on missing file returned error: %v", err)
	}
	rec, err = LoadPending(root)
	if err != nil || rec != nil {
		t.Fatalf("expected no pending, got %+v (err %v)", rec, err)
	}
}
