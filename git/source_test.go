package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/snapshot"
)

// testRepo creates an on-disk git repo with an initial config file committed on
// the master branch, and returns its path plus a commit helper.
func testRepo(t *testing.T, path, content string) (dir string, commit func(msg, content string) plumbing.Hash) {
	t.Helper()
	dir = t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	commit = func(msg, content string) plumbing.Hash {
		t.Helper()
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := wt.Add(path); err != nil {
			t.Fatal(err)
		}
		h, err := wt.Commit(msg, &gogit.CommitOptions{
			Author: &object.Signature{Name: "test", Email: "test@example.com", When: time.Now()},
		})
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	commit("initial", content)
	return dir, commit
}

type dbConfig struct {
	Host string `cfg:"host,git:server.host"`
	Port int    `cfg:"port,git:server.port"`
	Name string `cfg:"name"`
}

func TestGitSource_ResolvesFile(t *testing.T) {
	dir, _ := testRepo(t, "config.json", `{"server":{"host":"db.internal","port":5432},"name":"orders"}`)

	src, err := New(Options{Repo: dir, Ref: "master", Path: "config.json"})
	if err != nil {
		t.Fatal(err)
	}
	var cfg dbConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if cfg.Host != "db.internal" || cfg.Port != 5432 || cfg.Name != "orders" {
		t.Fatalf("got %+v, want {db.internal 5432 orders}", cfg)
	}
	if src.Version() == "" {
		t.Error("Version() empty after load")
	}
}

func TestGitSource_WatchDetectsDrift(t *testing.T) {
	dir, commit := testRepo(t, "config.json", `{"server":{"host":"old","port":1},"name":"svc"}`)
	src, err := New(Options{Repo: dir, Ref: "master", Path: "config.json", Poll: 30 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	var cfg dbConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 4)
	go src.Watch(ctx, func() { changed <- struct{}{} })

	// New commit on the same ref should be picked up as drift.
	commit("update host", `{"server":{"host":"new","port":2},"name":"svc"}`)

	select {
	case <-changed:
	case <-time.After(5 * time.Second):
		t.Fatal("Watch did not detect the new commit")
	}

	var reloaded dbConfig
	if _, err := gostructor.Configure(&reloaded, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}
	if reloaded.Host != "new" || reloaded.Port != 2 {
		t.Fatalf("after drift got %+v, want host=new port=2", reloaded)
	}
}

func TestGitSource_SetVersionSwitchesRef(t *testing.T) {
	dir, commit := testRepo(t, "config.json", `{"server":{"host":"v1host","port":1},"name":"svc"}`)
	// Tag the first commit, then move master forward.
	repo, _ := gogit.PlainOpen(dir)
	head, _ := repo.Head()
	if _, err := repo.CreateTag("v1", head.Hash(), nil); err != nil {
		t.Fatal(err)
	}
	commit("v2", `{"server":{"host":"v2host","port":2},"name":"svc"}`)

	src, err := New(Options{Repo: dir, Ref: "master", Path: "config.json"})
	if err != nil {
		t.Fatal(err)
	}
	var cfg dbConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src)); err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "v2host" {
		t.Fatalf("master host = %q, want v2host", cfg.Host)
	}

	if err := src.SetVersion("v1"); err != nil {
		t.Fatalf("SetVersion(v1): %v", err)
	}
	var pinned dbConfig
	if _, err := gostructor.Configure(&pinned, gostructor.WithSources(src)); err != nil {
		t.Fatal(err)
	}
	if pinned.Host != "v1host" {
		t.Fatalf("after SetVersion(v1) host = %q, want v1host", pinned.Host)
	}
}

func TestGitSource_SetVersionBadRefRollsBack(t *testing.T) {
	dir, _ := testRepo(t, "config.json", `{"server":{"host":"h","port":1},"name":"svc"}`)
	src, _ := New(Options{Repo: dir, Ref: "master", Path: "config.json"})
	var cfg dbConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src)); err != nil {
		t.Fatal(err)
	}
	if err := src.SetVersion("does-not-exist"); err == nil {
		t.Fatal("expected error switching to a nonexistent ref")
	}
	// Still serving the good ref.
	var after dbConfig
	if _, err := gostructor.Configure(&after, gostructor.WithSources(src)); err != nil {
		t.Fatal(err)
	}
	if after.Host != "h" {
		t.Fatalf("after failed switch host = %q, want h (rolled back)", after.Host)
	}
}

func TestGitSource_SnapshotFallbackWhenRepoUnreachable(t *testing.T) {
	store, err := snapshot.NewDirStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save("deadbeef", []byte(`{"server":{"host":"cached","port":9},"name":"svc"}`)); err != nil {
		t.Fatal(err)
	}
	src, err := New(Options{
		Repo:     filepath.Join(t.TempDir(), "nonexistent.git"),
		Ref:      "master",
		Path:     "config.json",
		Snapshot: store,
	})
	if err != nil {
		t.Fatal(err)
	}
	var cfg dbConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src)); err != nil {
		t.Fatalf("Configure should fall back to snapshot, got: %v", err)
	}
	if cfg.Host != "cached" || cfg.Port != 9 {
		t.Fatalf("snapshot fallback got %+v, want host=cached port=9", cfg)
	}
	if src.Version() != "deadbeef" {
		t.Errorf("Version() = %q, want deadbeef from snapshot", src.Version())
	}
}

func TestGitSource_SavesSnapshotOnGoodFetch(t *testing.T) {
	dir, _ := testRepo(t, "config.json", `{"server":{"host":"h","port":1},"name":"svc"}`)
	store, _ := snapshot.NewDirStore(t.TempDir())
	src, _ := New(Options{Repo: dir, Ref: "master", Path: "config.json", Snapshot: store})
	var cfg dbConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src)); err != nil {
		t.Fatal(err)
	}
	version, data, err := store.Load()
	if err != nil {
		t.Fatalf("snapshot not saved: %v", err)
	}
	if version == "" || len(data) == 0 {
		t.Fatalf("snapshot incomplete: version=%q len(data)=%d", version, len(data))
	}
}
