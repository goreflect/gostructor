package watch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goreflect/gostructor"
)

type fileConfig struct {
	Host  string `cfg:"host,file:server.host"`
	Port  int    `cfg:"port,file:server.port"`
	Level string `cfg:"level" gos:"default:info"`
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFileSource_Resolves(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeFile(t, path, `{"server":{"host":"localhost","port":8080}}`)

	src, err := JSONFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg fileConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "localhost" || cfg.Port != 8080 || cfg.Level != "info" {
		t.Fatalf("got %+v", cfg)
	}
}

func TestFileSource_WatchReloadsOnWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeFile(t, path, `{"server":{"host":"old","port":1}}`)

	src, err := JSONFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg fileConfig
	if _, err := gostructor.Configure(&cfg, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 4)
	go src.Watch(ctx, func() { changed <- struct{}{} })

	time.Sleep(100 * time.Millisecond) // let the watcher register
	writeFile(t, path, `{"server":{"host":"new","port":2}}`)

	select {
	case <-changed:
	case <-time.After(5 * time.Second):
		t.Fatal("Watch did not fire on file write")
	}

	var reloaded fileConfig
	if _, err := gostructor.Configure(&reloaded, gostructor.WithSources(src, gostructor.Default())); err != nil {
		t.Fatal(err)
	}
	if reloaded.Host != "new" || reloaded.Port != 2 {
		t.Fatalf("after write got %+v, want host=new port=2", reloaded)
	}
}

func TestFileSource_EndToEndWithGostructorWatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeFile(t, path, `{"server":{"host":"one","port":1}}`)

	src, err := JSONFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reloads := make(chan string, 8)
	go gostructor.Watch(ctx, &fileConfig{}, func(c *fileConfig, err error) {
		if err == nil {
			reloads <- c.Host
		}
	}, gostructor.WithSources(src, gostructor.Default()), gostructor.WithDebounce(50*time.Millisecond))

	if got := <-reloads; got != "one" { // initial fill
		t.Fatalf("initial = %q, want one", got)
	}

	time.Sleep(100 * time.Millisecond)
	writeFile(t, path, `{"server":{"host":"two","port":2}}`)

	select {
	case got := <-reloads:
		if got != "two" {
			t.Fatalf("reload = %q, want two", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no reload through gostructor.Watch")
	}
}
