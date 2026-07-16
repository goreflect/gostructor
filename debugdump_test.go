package gostructor

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

type dumpConfig struct {
	Host   string `cfg:"host,env:GOSTRUCTOR_DUMP_HOST" gos:"default:0.0.0.0"`
	Name   string `cfg:"name,env:GOSTRUCTOR_DUMP_NAME" gos:"optional"`
	Port   int    `cfg:"port,env:GOSTRUCTOR_DUMP_PORT" gos:"default:8080"`
	Secret string `cfg:"secret,env:GOSTRUCTOR_DUMP_SECRET" gos:"secret,optional"`
}

// readDumpFile returns path's contents (the write is synchronous with Configure).
func readDumpFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading dump file %s: %v", path, err)
	}
	return string(b)
}

// writeDumpJSON writes a JSON config file used as the primary source, so an
// env-sourced Port shows up in report.String() as a printed override (the
// report prints values only for overrides and secrets, not primary fields).
func writeDumpJSON(t *testing.T, body string) string {
	t.Helper()
	path := t.TempDir() + "/dump.json"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing json: %v", err)
	}
	return path
}

// dumpSources is Env over a JSON primary, the arrangement that makes an
// env-provided Port an override (printed) while JSON stays the primary source.
func dumpSources(jsonPath string) Option {
	return WithSources(Env(), JSONFile(jsonPath))
}

func TestDumpFileWritesReport(t *testing.T) {
	json := writeDumpJSON(t, `{"host":"db","name":"svc","port":8080}`)
	t.Setenv("GOSTRUCTOR_DUMP_PORT", "9090")
	path := t.TempDir() + "/dump.txt"

	if _, err := Configure(&dumpConfig{}, dumpSources(json), WithDebugDump(DumpFile(path))); err != nil {
		t.Fatalf("Configure: %v", err)
	}

	got := readDumpFile(t, path)
	if !strings.Contains(got, "gostructor config dump — gostructor.dumpConfig") {
		t.Errorf("dump file missing report header:\n%s", got)
	}
	if !strings.Contains(got, "9090") {
		t.Errorf("dump file missing overridden Port value:\n%s", got)
	}
}

func TestDumpFileRewrittenOnReconfigure(t *testing.T) {
	json := writeDumpJSON(t, `{"host":"db","name":"svc","port":8080}`)
	path := t.TempDir() + "/dump.txt"
	d := DumpFile(path)

	t.Setenv("GOSTRUCTOR_DUMP_PORT", "1111")
	if _, err := Configure(&dumpConfig{}, dumpSources(json), WithDebugDump(d)); err != nil {
		t.Fatalf("Configure #1: %v", err)
	}
	if got := readDumpFile(t, path); !strings.Contains(got, "1111") {
		t.Fatalf("first dump missing 1111:\n%s", got)
	}

	t.Setenv("GOSTRUCTOR_DUMP_PORT", "2222")
	if _, err := Configure(&dumpConfig{}, dumpSources(json), WithDebugDump(d)); err != nil {
		t.Fatalf("Configure #2: %v", err)
	}
	got := readDumpFile(t, path)
	if strings.Contains(got, "1111") || !strings.Contains(got, "2222") {
		t.Errorf("dump not replaced on reconfigure, want 2222 not 1111:\n%s", got)
	}
}

func TestDumpFileMasksSecret(t *testing.T) {
	t.Setenv("GOSTRUCTOR_DUMP_SECRET", "supersecret")
	path := t.TempDir() + "/dump.txt"

	if _, err := Configure(&dumpConfig{}, WithDebugDump(DumpFile(path))); err != nil {
		t.Fatalf("Configure: %v", err)
	}

	got := readDumpFile(t, path)
	if strings.Contains(got, "supersecret") {
		t.Fatalf("secret leaked into dump file:\n%s", got)
	}
	if !strings.Contains(got, "••••••") {
		t.Errorf("expected masked secret in dump file:\n%s", got)
	}
}

// dialDump reads one report from a tcpDumper's live listener.
func dialDump(t *testing.T, addr string) string {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dialing dump endpoint %s: %v", addr, err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var b strings.Builder
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		b.WriteString(sc.Text())
		b.WriteString("\n")
	}
	return b.String()
}

func TestDumpTCPServesReport(t *testing.T) {
	json := writeDumpJSON(t, `{"host":"db","name":"svc","port":8080}`)
	t.Setenv("GOSTRUCTOR_DUMP_PORT", "7777")
	// Bind an ephemeral port so parallel runs never collide.
	d := &tcpDumper{addr: "127.0.0.1:0"}

	if _, err := Configure(&dumpConfig{}, dumpSources(json), WithDebugDump(d)); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	defer d.Close()

	d.mu.Lock()
	ln := d.ln
	d.mu.Unlock()
	if ln == nil {
		t.Fatal("listener did not start after Configure")
	}

	got := dialDump(t, ln.Addr().String())
	if !strings.Contains(got, "gostructor config dump — gostructor.dumpConfig") {
		t.Errorf("served report missing header:\n%s", got)
	}
	if !strings.Contains(got, "7777") {
		t.Errorf("served report missing overridden Port value:\n%s", got)
	}
}

func TestDumpTCPCloseStopsListener(t *testing.T) {
	d := &tcpDumper{addr: "127.0.0.1:0"}
	if _, err := Configure(&dumpConfig{}, WithDebugDump(d)); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	d.mu.Lock()
	addr := d.ln.Addr().String()
	d.mu.Unlock()

	if err := d.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := net.DialTimeout("tcp", addr, 500*time.Millisecond); err == nil {
		t.Error("expected dial to fail after Close, but it succeeded")
	}
}

func TestDumpTCPBindFailureIsNonFatal(t *testing.T) {
	// Hold a port, then point a dumper at it: Configure must still succeed.
	held, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("holding port: %v", err)
	}
	defer held.Close()

	d := &tcpDumper{addr: held.Addr().String()}
	if _, err := Configure(&dumpConfig{}, WithDebugDump(d)); err != nil {
		t.Fatalf("Configure must not fail on a busy dump port, got: %v", err)
	}
}

func TestWithDebugDumpNilIsInert(t *testing.T) {
	if _, err := Configure(&dumpConfig{}, WithDebugDump(nil)); err != nil {
		t.Fatalf("Configure with nil dumper: %v", err)
	}
}

func TestNoDumperByDefault(t *testing.T) {
	t.Setenv(EnvDebugDump, "")
	c := newConfig(nil)
	if c.dumper != nil {
		t.Errorf("expected no dumper by default, got %T", c.dumper)
	}
}

func TestDumperFromEnv(t *testing.T) {
	cases := []struct {
		env  string
		want any // nil, *tcpDumper, or *fileDumper
		addr string
	}{
		{env: "", want: nil},
		{env: "off", want: nil},
		{env: "false", want: nil},
		{env: "on", want: (*tcpDumper)(nil), addr: defaultDumpAddr},
		{env: "tcp", want: (*tcpDumper)(nil), addr: defaultDumpAddr},
		{env: "127.0.0.1:9999", want: (*tcpDumper)(nil), addr: "127.0.0.1:9999"},
		{env: "file:/tmp/x.txt", want: (*fileDumper)(nil)},
	}
	for _, tc := range cases {
		t.Run(tc.env, func(t *testing.T) {
			t.Setenv(EnvDebugDump, tc.env)
			c := newConfig(nil)
			switch tc.want.(type) {
			case nil:
				if c.dumper != nil {
					t.Fatalf("env %q: want nil dumper, got %T", tc.env, c.dumper)
				}
			case *tcpDumper:
				td, ok := c.dumper.(*tcpDumper)
				if !ok {
					t.Fatalf("env %q: want *tcpDumper, got %T", tc.env, c.dumper)
				}
				if td.addr != tc.addr {
					t.Errorf("env %q: addr = %q, want %q", tc.env, td.addr, tc.addr)
				}
			case *fileDumper:
				if _, ok := c.dumper.(*fileDumper); !ok {
					t.Fatalf("env %q: want *fileDumper, got %T", tc.env, c.dumper)
				}
			}
		})
	}
}

func TestWatchFeedsDumperOnReloadAndClosesOnCancel(t *testing.T) {
	src := newFakeWatchable("one")
	d := &tcpDumper{addr: "127.0.0.1:0"}

	ctx, cancel := context.WithCancel(context.Background())
	reloaded := make(chan struct{}, 8)
	onReload := func(_ *watchConfig, err error) {
		if err != nil {
			t.Errorf("reload error: %v", err)
		}
		reloaded <- struct{}{}
	}

	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, &watchConfig{}, onReload,
			WithSources(src, Default()), WithDebugDump(d))
	}()

	<-reloaded  // initial fill, which starts the listener
	<-src.ready // subscribed

	d.mu.Lock()
	ln := d.ln
	d.mu.Unlock()
	if ln == nil {
		t.Fatal("listener did not start on initial fill")
	}
	addr := ln.Addr().String()
	if got := dialDump(t, addr); !strings.Contains(got, "one") {
		t.Errorf("initial dump missing value 'one':\n%s", got)
	}

	// A reload must refresh what the endpoint serves.
	src.set("two")
	<-reloaded
	if got := dialDump(t, addr); !strings.Contains(got, "two") {
		t.Errorf("dump not refreshed after reload, missing 'two':\n%s", got)
	}

	// Cancelling the watch must close the listener.
	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatalf("Watch returned %v, want context.Canceled", err)
	}
	if _, err := net.DialTimeout("tcp", addr, 500*time.Millisecond); err == nil {
		t.Error("expected listener closed after Watch ended, but dial succeeded")
	}
}

func TestWatchRejectedReloadKeepsDump(t *testing.T) {
	src := newFakeWatchable("good")
	d := &tcpDumper{addr: "127.0.0.1:0"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reloaded := make(chan struct{}, 8)
	rejected := make(chan struct{}, 8)
	onReload := func(_ *watchConfig, err error) {
		if err != nil {
			rejected <- struct{}{}
			return
		}
		reloaded <- struct{}{}
	}

	go Watch(ctx, &watchConfig{}, onReload,
		WithSources(src, Default()),
		WithValidate(func(c *watchConfig) error {
			if c.Value == "bad" {
				return fmt.Errorf("value %q rejected", c.Value)
			}
			return nil
		}),
		WithDebugDump(d),
	)

	<-reloaded  // initial "good" fill, listener up
	<-src.ready // subscribed
	d.mu.Lock()
	addr := d.ln.Addr().String()
	d.mu.Unlock()

	// A reload that fails validation must not touch the dump.
	src.set("bad")
	<-rejected
	got := dialDump(t, addr)
	if strings.Contains(got, "bad") {
		t.Errorf("rejected reload leaked into dump:\n%s", got)
	}
	if !strings.Contains(got, "good") {
		t.Errorf("dump should still show last-known-good 'good':\n%s", got)
	}
}

func TestExplicitOptionBeatsEnv(t *testing.T) {
	// Env says on, but code explicitly disables via nil: option must win.
	t.Setenv(EnvDebugDump, "on")
	c := newConfig([]Option{WithDebugDump(nil)})
	if c.dumper != nil {
		t.Errorf("explicit WithDebugDump(nil) should override env, got %T", c.dumper)
	}
}
