// Package watch provides a file-backed gostructor.Source that reloads when the
// file changes on disk, using fsnotify. It is the local, dependency-light way to
// get hot reload: point a service at a JSON (or, with a custom decoder, any)
// config file and have gostructor.Watch re-fill the struct on every save.
//
// It watches the file's *directory* rather than the file inode, because editors
// and atomic writers replace a file by renaming a new one over it — which drops
// a watch on the old inode. Directory watching survives that, and the source
// filters events down to the target file.
//
// This lives in its own module so the core stays dependency-free; only programs
// that want file watching pull in fsnotify.
package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/goreflect/gostructor"
)

// DefaultName is the source identity and cfg per-source override key.
const DefaultName = "file"

// Decoder turns the file's bytes into a nested map addressable by
// gostructor.LookupKey. The default decodes JSON.
type Decoder func([]byte) (map[string]any, error)

// Options configures a watched file source.
type Options struct {
	// Path is the config file to read and watch. Required.
	Path string
	// Decoder parses the file bytes; nil defaults to JSON.
	Decoder Decoder
	// Name overrides the source identity (DefaultName otherwise).
	Name string
	// Logger receives watch diagnostics; nil discards them.
	Logger *slog.Logger
}

type source struct {
	path   string
	name   string
	decode Decoder
	log    *slog.Logger

	mu      sync.Mutex
	data    map[string]any
	loaded  bool
	loadErr error
}

// New returns a file-backed gostructor.Source that also implements
// gostructor.Watchable. The file is read lazily on first Resolve.
func New(opts Options) (*source, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("gostructor/watch: Path is required")
	}
	abs, err := filepath.Abs(opts.Path)
	if err != nil {
		return nil, fmt.Errorf("gostructor/watch: resolving %q: %w", opts.Path, err)
	}
	s := &source{
		path:   abs,
		name:   opts.Name,
		decode: opts.Decoder,
		log:    opts.Logger,
	}
	if s.name == "" {
		s.name = DefaultName
	}
	if s.decode == nil {
		s.decode = jsonDecoder
	}
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}
	return s, nil
}

// JSONFile is a shortcut for New(Options{Path: path}) — a watched JSON file
// under the default source name.
func JSONFile(path string) (*source, error) {
	return New(Options{Path: path})
}

func (s *source) Name() string { return s.name }

// Resolve fills a field from the file, reading it on first use.
func (s *source) Resolve(field gostructor.FieldContext) (any, bool, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, false, err
	}
	s.mu.Lock()
	data := s.data
	s.mu.Unlock()
	value, ok := gostructor.LookupKey(field, s.name, data)
	return value, ok, nil
}

// Watch reports file changes via fsnotify, calling onChange after re-reading the
// file so the reload sees the new contents. It watches the containing directory
// so it survives atomic (rename-over) writes. It blocks until ctx is cancelled.
func (s *source) Watch(ctx context.Context, onChange func()) error {
	if err := s.ensureLoaded(); err != nil {
		s.log.Warn("gostructor/watch: initial read failed, will retry on events", "err", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("gostructor/watch: creating watcher: %w", err)
	}
	defer watcher.Close()

	dir := filepath.Dir(s.path)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("gostructor/watch: watching %q: %w", dir, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if filepath.Clean(event.Name) != s.path {
				continue // some other file in the directory
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			if err := s.reload(); err != nil {
				s.log.Warn("gostructor/watch: reread after change failed", "err", err)
				continue
			}
			onChange()
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			s.log.Warn("gostructor/watch: watcher error", "err", err)
		}
	}
}

func (s *source) ensureLoaded() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded {
		return s.loadErr
	}
	s.loaded = true
	data, err := s.read()
	if err != nil {
		s.loadErr = err
		return err
	}
	s.data = data
	return nil
}

func (s *source) reload() error {
	data, err := s.read()
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.data = data
	s.mu.Unlock()
	return nil
}

func (s *source) read() (map[string]any, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("gostructor/watch: reading %q: %w", s.path, err)
	}
	data, err := s.decode(raw)
	if err != nil {
		return nil, fmt.Errorf("gostructor/watch: decoding %q: %w", s.path, err)
	}
	return data, nil
}

func jsonDecoder(raw []byte) (map[string]any, error) {
	parsed := map[string]any{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}
