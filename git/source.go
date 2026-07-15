// Package git resolves gostructor fields from a config file kept in a git
// repository, treating a branch or tag as a config version. It clones the repo
// in memory, reads one file at the chosen ref, and addresses fields by nested
// key like the JSON source.
//
// It implements gostructor.Watchable: it polls the ref for a new commit and
// signals a reload when the SHA moves, so a service follows the ref without a
// restart; SetVersion re-points to another ref at runtime. With a
// gostructor/snapshot store, every good fetch is persisted and served as
// last-known-good when the remote is unreachable at startup.
package git

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/snapshot"
)

// DefaultName is this source's identity and its cfg per-source override key,
// e.g. `cfg:"host,git:server.host"`. Override it via Options.Name.
const DefaultName = "git"

// DefaultPoll is the drift-check interval used when Options.Poll is zero.
const DefaultPoll = 30 * time.Second

// Decoder parses the config file bytes into a nested map. The default is JSON;
// pass another via Options.Decoder (gostructor.DecodeINI/DecodeKeyValue, or
// yaml.Decode/toml.Decode/hocon.Decode).
type Decoder = gostructor.Decoder

// Options configures a git Source.
type Options struct {
	// Repo is the repository URL: an https/ssh remote, or a local path/URL for
	// tests and on-disk mirrors. Required.
	Repo string
	// Ref is the branch or tag that is the config version, e.g. "main" or
	// "release/2025.10". Empty defaults to "HEAD" (the remote's default branch).
	Ref string
	// Path is the config file's path within the repo, e.g.
	// "services/api/config.json". Required.
	Path string
	// Poll is how often Watch re-checks the ref for a new commit. Zero uses
	// DefaultPoll; a negative value disables polling (Watch then blocks until
	// its context is cancelled, and only SetVersion triggers a reload).
	Poll time.Duration
	// Auth is the optional git transport auth (a token, ssh key). Nil relies on
	// anonymous access, which is enough for a public repo over https.
	Auth transport.AuthMethod
	// Snapshot, if set, receives every good fetch and is read back when the
	// remote is unreachable at startup, so a git outage doesn't take the app
	// down. See gostructor/snapshot.
	Snapshot snapshot.Store
	// Decoder parses the config file bytes; nil defaults to JSON.
	Decoder Decoder
	// Name overrides the source identity (DefaultName otherwise).
	Name string
	// Logger receives drift/fallback diagnostics; nil discards them.
	Logger *slog.Logger
}

type source struct {
	opts   Options
	name   string
	decode Decoder
	log    *slog.Logger

	mu       sync.Mutex
	repo     *gogit.Repository
	fs       *memory.Storage
	ref      string
	version  string // resolved commit SHA of the last good read
	data     map[string]any
	loaded   bool
	loadErr  error
	onChange func()
}

// New returns a git-backed gostructor.Source (also a gostructor.Watchable). The
// repo is cloned lazily on first Resolve, so constructing a source never blocks
// or fails; a bad repo/path surfaces as a resolution error (or a snapshot
// fallback) at fill time.
func New(opts Options) (*source, error) {
	if opts.Repo == "" {
		return nil, fmt.Errorf("gostructor/git: Repo is required")
	}
	if opts.Path == "" {
		return nil, fmt.Errorf("gostructor/git: Path is required")
	}
	s := &source{
		opts:   opts,
		name:   opts.Name,
		decode: opts.Decoder,
		log:    opts.Logger,
		ref:    opts.Ref,
	}
	if s.name == "" {
		s.name = DefaultName
	}
	if s.decode == nil {
		s.decode = gostructor.DecodeJSON
	}
	if s.log == nil {
		s.log = slog.New(slog.DiscardHandler)
	}
	if s.ref == "" {
		s.ref = "HEAD"
	}
	return s, nil
}

func (s *source) Name() string { return s.name }

// Resolve fills a field from the file at the current ref, loading the repo on
// first use.
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

// Version returns the commit SHA the currently served config was read from,
// useful for logging provenance. It is empty until the first successful load.
func (s *source) Version() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.version
}

// SetVersion re-points the source at another branch/tag/ref at runtime, re-reads
// the config file there, and — if a Watch is running — triggers a reload. The
// currently served config stays in place until the new ref is fully read, so a
// bad ref name returns an error and changes nothing.
func (s *source) SetVersion(ref string) error {
	s.mu.Lock()
	prevRef, prevVersion := s.ref, s.version
	s.ref = ref
	changed, err := s.syncLocked(true)
	onChange := s.onChange
	if err != nil {
		s.ref, s.version = prevRef, prevVersion // roll back the re-point
		s.mu.Unlock()
		return fmt.Errorf("gostructor/git: switching to ref %q: %w", ref, err)
	}
	s.mu.Unlock()
	if changed && onChange != nil {
		onChange()
	}
	return nil
}

// Watch re-checks the ref for a new commit on the poll interval and calls
// onChange when the SHA moves, after refreshing the in-memory config so the
// reload reads the new bytes. It blocks until ctx is cancelled.
func (s *source) Watch(ctx context.Context, onChange func()) error {
	// Make sure the repo is loaded so the first poll has a baseline SHA.
	if err := s.ensureLoaded(); err != nil {
		s.log.Warn("gostructor/git: initial load failed, watch will keep retrying", "err", err)
	}
	s.mu.Lock()
	s.onChange = onChange
	poll := s.opts.Poll
	s.mu.Unlock()

	if poll < 0 {
		<-ctx.Done()
		return ctx.Err()
	}
	if poll == 0 {
		poll = DefaultPoll
	}
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.mu.Lock()
			changed, err := s.syncLocked(true)
			s.mu.Unlock()
			if err != nil {
				s.log.Warn("gostructor/git: drift check failed, keeping current config", "err", err)
				continue
			}
			if changed {
				s.log.Debug("gostructor/git: ref advanced, reloading", "ref", s.ref, "version", s.Version())
				onChange()
			}
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
	if _, err := s.syncLocked(false); err != nil {
		if s.loadFromSnapshotLocked(err) {
			return nil
		}
		s.loadErr = err
	}
	return s.loadErr
}

// syncLocked clones (first call) or fetches (fetch=true) the repo, resolves the
// current ref to a commit, and — when that commit differs from what is loaded —
// re-reads and parses the config file, updating data/version and saving a
// snapshot. It reports whether the served config changed. The caller holds mu.
func (s *source) syncLocked(fetch bool) (changed bool, err error) {
	if s.repo == nil {
		if err := s.cloneLocked(); err != nil {
			return false, err
		}
	} else if fetch {
		if err := s.fetchLocked(); err != nil {
			return false, err
		}
	}

	hash, err := s.resolveRefLocked(s.ref)
	if err != nil {
		return false, err
	}
	if hash.String() == s.version && s.data != nil {
		return false, nil
	}

	raw, err := s.readFileLocked(hash)
	if err != nil {
		return false, err
	}
	data, err := s.decode(raw)
	if err != nil {
		return false, fmt.Errorf("gostructor/git: decoding %q at %s: %w", s.opts.Path, hash.String()[:min(7, len(hash.String()))], err)
	}
	s.version = hash.String()
	s.data = data
	if s.opts.Snapshot != nil {
		if err := s.opts.Snapshot.Save(s.version, raw); err != nil {
			s.log.Warn("gostructor/git: saving snapshot failed", "err", err)
		}
	}
	return true, nil
}

func (s *source) cloneLocked() error {
	s.fs = memory.NewStorage()
	repo, err := gogit.Clone(s.fs, memfs.New(), &gogit.CloneOptions{
		URL:  s.opts.Repo,
		Auth: s.opts.Auth,
		Tags: gogit.AllTags,
	})
	if err != nil {
		s.fs = nil
		return fmt.Errorf("gostructor/git: cloning %q: %w", s.opts.Repo, err)
	}
	s.repo = repo
	return nil
}

func (s *source) fetchLocked() error {
	err := s.repo.Fetch(&gogit.FetchOptions{
		Auth:  s.opts.Auth,
		Tags:  gogit.AllTags,
		Force: true,
	})
	if err != nil && !errors.Is(err, gogit.NoErrAlreadyUpToDate) {
		return fmt.Errorf("gostructor/git: fetching %q: %w", s.opts.Repo, err)
	}
	return nil
}

// resolveRefLocked maps a user ref ("main", "v1.2", a SHA, "HEAD") to a commit
// hash. It tries the remote-tracking form first so a branch follows origin
// (which fetch advances) rather than the stale local branch — that ordering is
// what makes drift detection work. Tags and raw SHAs fall through.
func (s *source) resolveRefLocked(ref string) (plumbing.Hash, error) {
	var candidates []string
	if ref == "HEAD" {
		candidates = []string{"refs/remotes/origin/HEAD", "HEAD"}
	} else {
		candidates = []string{
			"refs/remotes/origin/" + ref,
			"origin/" + ref,
			"refs/tags/" + ref,
			ref,
		}
	}
	var lastErr error
	for _, c := range candidates {
		h, err := s.repo.ResolveRevision(plumbing.Revision(c))
		if err == nil && h != nil {
			return *h, nil
		}
		lastErr = err
	}
	return plumbing.ZeroHash, fmt.Errorf("gostructor/git: cannot resolve ref %q: %w", ref, lastErr)
}

func (s *source) readFileLocked(hash plumbing.Hash) ([]byte, error) {
	commit, err := s.repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("gostructor/git: reading commit %s: %w", hash, err)
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("gostructor/git: reading tree at %s: %w", hash, err)
	}
	file, err := tree.File(s.opts.Path)
	if err != nil {
		if errors.Is(err, object.ErrFileNotFound) {
			return nil, fmt.Errorf("gostructor/git: file %q not found at %s", s.opts.Path, hash)
		}
		return nil, fmt.Errorf("gostructor/git: locating %q at %s: %w", s.opts.Path, hash, err)
	}
	reader, err := file.Reader()
	if err != nil {
		return nil, fmt.Errorf("gostructor/git: opening %q: %w", s.opts.Path, err)
	}
	defer func() { _ = reader.Close() }()
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("gostructor/git: reading %q: %w", s.opts.Path, err)
	}
	return raw, nil
}

// loadFromSnapshotLocked tries to serve the last-known-good snapshot after a
// failed clone/fetch, so the app starts even when git is unreachable. It reports
// whether a snapshot was successfully loaded. cause is the upstream error, kept
// for the log.
func (s *source) loadFromSnapshotLocked(cause error) bool {
	if s.opts.Snapshot == nil {
		return false
	}
	version, raw, err := s.opts.Snapshot.Load()
	if err != nil {
		return false
	}
	data, err := s.decode(raw)
	if err != nil {
		s.log.Warn("gostructor/git: stored snapshot is undecodable", "err", err)
		return false
	}
	s.version = version
	s.data = data
	s.log.Warn("gostructor/git: upstream unavailable, serving last-known-good snapshot",
		"cause", cause, "version", version)
	return true
}
