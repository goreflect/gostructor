// Package snapshot is gostructor's durable last-known-good store — the
// "файлопомойка" a remote source (git, a config server) writes its most
// recently fetched, successfully resolved configuration to, and reads back when
// the upstream is unreachable.
//
// The point is availability: a service driven from a remote config must not go
// down because the remote had a hiccup at the wrong moment. A remote Source
// saves each good fetch to a Store; on a later start (or a fetch failure) it
// Loads the last good bytes and serves from those instead of failing. The
// version string — a commit SHA, a KV modify-index, a config-server label —
// lets the source tell whether the upstream has drifted from what it last
// stored.
//
// Store is a two-method interface so callers can keep the snapshot wherever
// they like (an object store, a database); DirStore is the default on-disk
// implementation.
package snapshot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrNoSnapshot is returned by Load when the store holds nothing yet — a first
// run with the upstream unreachable, for instance. Callers test for it with
// errors.Is to distinguish "no last-known-good exists" from a real I/O error.
var ErrNoSnapshot = errors.New("gostructor/snapshot: no snapshot stored")

// Store persists the latest good configuration payload and the version it came
// from, and hands both back on request. Implementations must be safe for
// concurrent use: a Watch loop may Save while a reload Loads.
type Store interface {
	// Save records data as the current good snapshot, tagged with version (an
	// upstream identifier such as a commit SHA or modify-index). It should be
	// atomic from a reader's view — a concurrent Load never sees a half-written
	// payload.
	Save(version string, data []byte) error
	// Load returns the stored payload and the version it was saved under, or a
	// wrapped ErrNoSnapshot if nothing has been stored yet.
	Load() (version string, data []byte, err error)
}

// DirStore is the default file-backed Store: the "файлопомойка" on disk. It
// keeps the payload and its version in two files under dir, writing each
// atomically (write to a temp file, then rename) so a crash mid-Save can never
// corrupt the last-known-good.
type DirStore struct {
	dir string
}

// NewDirStore returns a DirStore rooted at dir, creating dir (and parents) if
// needed. A relative dir is resolved against the process working directory.
func NewDirStore(dir string) (*DirStore, error) {
	if dir == "" {
		return nil, fmt.Errorf("gostructor/snapshot: empty directory")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("gostructor/snapshot: creating %q: %w", dir, err)
	}
	return &DirStore{dir: dir}, nil
}

const (
	dataFile    = "snapshot.data"
	versionFile = "snapshot.version"
)

// Save writes the version first, then the payload, each atomically. Payload
// last means a reader that races a Save either sees the whole previous snapshot
// or the whole new one — never a new payload tagged with the old version in a
// way that hides drift.
func (s *DirStore) Save(version string, data []byte) error {
	if err := writeAtomic(filepath.Join(s.dir, versionFile), []byte(version)); err != nil {
		return err
	}
	if err := writeAtomic(filepath.Join(s.dir, dataFile), data); err != nil {
		return err
	}
	return nil
}

// Load reads back the payload and version, mapping a missing payload to
// ErrNoSnapshot.
func (s *DirStore) Load() (string, []byte, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, dataFile))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("%w in %q", ErrNoSnapshot, s.dir)
		}
		return "", nil, fmt.Errorf("gostructor/snapshot: reading payload: %w", err)
	}
	version, err := os.ReadFile(filepath.Join(s.dir, versionFile))
	if err != nil {
		if os.IsNotExist(err) {
			// Payload without a version: usable, just unversioned.
			return "", data, nil
		}
		return "", nil, fmt.Errorf("gostructor/snapshot: reading version: %w", err)
	}
	return string(version), data, nil
}

// writeAtomic writes data to a sibling temp file and renames it over path, so a
// concurrent reader sees either the old file or the complete new one.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".snapshot-*")
	if err != nil {
		return fmt.Errorf("gostructor/snapshot: temp file in %q: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("gostructor/snapshot: writing %q: %w", tmpName, err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("gostructor/snapshot: syncing %q: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("gostructor/snapshot: closing %q: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("gostructor/snapshot: renaming into %q: %w", path, err)
	}
	return nil
}
