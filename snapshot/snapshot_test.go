package snapshot

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestDirStore_SaveLoadRoundTrip(t *testing.T) {
	store, err := NewDirStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"host":"db.internal","port":5432}`)
	if err := store.Save("abc123", payload); err != nil {
		t.Fatalf("Save: %v", err)
	}
	version, data, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if version != "abc123" {
		t.Errorf("version = %q, want abc123", version)
	}
	if !bytes.Equal(data, payload) {
		t.Errorf("data = %q, want %q", data, payload)
	}
}

func TestDirStore_LoadMissingIsErrNoSnapshot(t *testing.T) {
	store, err := NewDirStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Load(); !errors.Is(err, ErrNoSnapshot) {
		t.Fatalf("Load on empty store = %v, want ErrNoSnapshot", err)
	}
}

func TestDirStore_OverwriteKeepsLatest(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewDirStore(dir)
	if err := store.Save("v1", []byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := store.Save("v2", []byte("second")); err != nil {
		t.Fatal(err)
	}
	version, data, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if version != "v2" || string(data) != "second" {
		t.Fatalf("after overwrite got %q/%q, want v2/second", version, data)
	}
	// Only the two known files plus no leftover temp files.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != dataFile && e.Name() != versionFile {
			t.Errorf("unexpected leftover file %q", e.Name())
		}
	}
}

func TestNewDirStore_CreatesNestedDir(t *testing.T) {
	nested := filepath.Join(t.TempDir(), "a", "b", "c")
	if _, err := NewDirStore(nested); err != nil {
		t.Fatalf("NewDirStore nested: %v", err)
	}
	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("nested dir not created: %v", err)
	}
}

func TestDirStore_ConcurrentSaveLoad(t *testing.T) {
	store, _ := NewDirStore(t.TempDir())
	if err := store.Save("v0", []byte("payload-0")); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); _ = store.Save("vN", []byte("payload-N")) }()
		go func() {
			defer wg.Done()
			// Load must always return a complete, non-empty payload.
			if _, data, err := store.Load(); err == nil && len(data) == 0 {
				t.Error("Load observed a torn (empty) snapshot")
			}
		}()
	}
	wg.Wait()
}
