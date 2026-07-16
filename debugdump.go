package gostructor

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// EnvDebugDump names the environment variable that turns the debug dump on
// without a code change, for inspecting a running service (typically in prod).
// It is consulted only when no WithDebugDump option is passed:
//
//	off | false | 0 | no   → disabled (same as unset)
//	on | tcp | 1           → DumpTCP on the default loopback address
//	host:port              → DumpTCP on that address
//	file:/path/to/file     → DumpFile at that path
const EnvDebugDump = "GOSTRUCTOR_DEBUG_DUMP"

// defaultDumpAddr is the loopback address DumpTCP binds when given "" and the
// address EnvDebugDump uses for its "on"/"tcp" shorthand. Loopback-only so a
// pod's debug port is reachable via kubectl exec / port-forward but not exposed
// on the pod network by default; pass an explicit host to widen it.
const defaultDumpAddr = "127.0.0.1:6555"

// Dumper receives the latest resolution Report after every successful fill — the
// initial Configure and each live reload under Watch — and exposes it for
// out-of-band inspection: a debug file to cat, a TCP port to nc. It is a
// debugging aid only. The report it is fed already has secret fields masked, and
// a Dumper never affects resolution: a sink failure (port busy, unwritable file)
// is logged and swallowed, never surfaced as a Configure error.
//
// A Dumper is opt-in via WithDebugDump (or EnvDebugDump); there is no default.
// A Dumper that also implements io.Closer is closed when Watch's context is
// cancelled; a one-shot Configure does not close its dumper, leaving the
// endpoint serving for the life of the process.
type Dumper interface {
	// Update publishes the latest report. It must not block or panic.
	Update(rep *Report)
}

// loggerAware lets the built-in dumpers receive Configure's logger so their
// operational warnings land where the caller's other gostructor logs do. A
// user-supplied Dumper need not implement it.
type loggerAware interface{ setLogger(*slog.Logger) }

// attachLogger wires cfg's logger into a built-in dumper.
func attachLogger(d Dumper, l *slog.Logger) {
	if la, ok := d.(loggerAware); ok {
		la.setLogger(l)
	}
}

// WithDebugDump publishes the resolution report to d after every successful
// fill, so a running service's effective configuration can be inspected out of
// band. Pass DumpTCP(...) to serve it on a port, DumpFile(...) to write it to
// disk, or your own Dumper. It is off unless set (or unless EnvDebugDump is
// set); passing a nil Dumper is the same as not passing the option.
func WithDebugDump(d Dumper) Option {
	return func(c *config) {
		c.dumper = d
		c.dumperSet = true
	}
}

// dumperFromEnv builds the dumper named by EnvDebugDump, or nil when the
// variable is unset or explicitly disabled. It is consulted only when no
// WithDebugDump option was passed.
func dumperFromEnv() Dumper {
	switch v := strings.TrimSpace(os.Getenv(EnvDebugDump)); strings.ToLower(v) {
	case "", "off", "false", "0", "no":
		return nil
	case "on", "tcp", "1":
		return DumpTCP(defaultDumpAddr)
	default:
		if path, ok := strings.CutPrefix(v, "file:"); ok {
			return DumpFile(path)
		}
		return DumpTCP(v)
	}
}

// DumpFile writes the full per-field resolution dump to path after every
// successful fill, replacing it in place. Reads never see a half-written file:
// the report is written to a sibling temp file and renamed over path. A write
// error is logged, not fatal. An empty path disables the dumper.
func DumpFile(path string) Dumper {
	if path == "" {
		return nil
	}
	return &fileDumper{path: path}
}

// DumpTCP serves the full per-field resolution dump on addr: each accepted
// connection receives the latest dump and is closed, so `nc host port` or the
// gostructor-dump client prints the live config with no HTTP involved. The
// listener starts on the first report and, if addr is already in use, is logged
// and disabled rather than failing Configure. An empty addr binds the default
// loopback address (127.0.0.1:6555).
func DumpTCP(addr string) Dumper {
	if addr == "" {
		addr = defaultDumpAddr
	}
	return &tcpDumper{addr: addr}
}

// dumpString renders the full per-field state of a resolution report: every
// field, its type, the source that won it (or its outcome when none did), and
// its value. Secret values are already masked in the report. Unlike
// report.String() — a focused override/secret summary for startup logs — this
// lists every field, which is what an operator inspecting a live service wants.
func dumpString(r *Report) string {
	nameW, typeW, srcW := len("FIELD"), len("TYPE"), len("SOURCE")
	for _, f := range r.Fields {
		nameW = max(nameW, len(f.Field))
		typeW = max(typeW, len(f.Type))
		srcW = max(srcW, len(dumpSource(f)))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "gostructor config dump — %s (%s)\n\n", orUnknownType(r.Type), pluralFields(len(r.Fields)))
	fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %s\n", nameW, "FIELD", typeW, "TYPE", srcW, "SOURCE", "VALUE")
	for _, f := range r.Fields {
		fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %s\n", nameW, f.Field, typeW, f.Type, srcW, dumpSource(f), dumpValue(f))
	}
	return b.String()
}

// dumpSource names the winning source, or the outcome in parentheses when no
// source produced a value (e.g. "(unresolved)").
func dumpSource(f FieldResolution) string {
	if f.Winner != "" {
		return f.Winner
	}
	return "(" + f.Outcome + ")"
}

// dumpValue renders a field's resolved value, or a dash when none was set.
func dumpValue(f FieldResolution) string {
	if f.Winner == "" {
		return "—"
	}
	return fmt.Sprint(f.Value)
}

// fileDumper is the DumpFile sink.
type fileDumper struct {
	path   string
	logger atomic.Pointer[slog.Logger]
}

func (d *fileDumper) setLogger(l *slog.Logger) { d.logger.Store(l) }

func (d *fileDumper) log() *slog.Logger {
	if l := d.logger.Load(); l != nil {
		return l
	}
	return slog.New(slog.DiscardHandler)
}

func (d *fileDumper) Update(rep *Report) {
	tmp := d.path + ".tmp"
	if err := os.WriteFile(tmp, []byte(dumpString(rep)), 0o644); err != nil {
		d.log().Warn("gostructor: debug dump file write failed", "path", d.path, "err", err)
		return
	}
	if err := os.Rename(tmp, d.path); err != nil {
		d.log().Warn("gostructor: debug dump file rename failed", "path", d.path, "err", err)
		_ = os.Remove(tmp)
	}
}

// tcpDumper is the DumpTCP sink. The listener starts once, on the first Update,
// and every accepted connection is served the latest report held in latest.
type tcpDumper struct {
	addr   string
	logger atomic.Pointer[slog.Logger]

	latest  atomic.Pointer[Report]
	start   sync.Once
	mu      sync.Mutex // guards ln
	ln      net.Listener
	stopped atomic.Bool
}

func (d *tcpDumper) setLogger(l *slog.Logger) { d.logger.Store(l) }

func (d *tcpDumper) log() *slog.Logger {
	if l := d.logger.Load(); l != nil {
		return l
	}
	return slog.New(slog.DiscardHandler)
}

func (d *tcpDumper) Update(rep *Report) {
	d.latest.Store(rep)
	d.start.Do(d.listen)
}

// listen binds the address and spawns the accept loop. A bind failure disables
// the dumper (the report is still stored, harmlessly) rather than failing the
// fill that triggered it.
func (d *tcpDumper) listen() {
	if d.stopped.Load() {
		return
	}
	ln, err := net.Listen("tcp", d.addr)
	if err != nil {
		d.log().Warn("gostructor: debug dump listener disabled", "addr", d.addr, "err", err)
		return
	}
	d.mu.Lock()
	// A Close that raced ahead of the first Update wins: honor it.
	if d.stopped.Load() {
		d.mu.Unlock()
		_ = ln.Close()
		return
	}
	d.ln = ln
	d.mu.Unlock()
	d.log().Debug("gostructor: debug dump listening", "addr", ln.Addr().String())
	go d.serve(ln)
}

func (d *tcpDumper) serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return // listener closed
		}
		go d.handle(conn)
	}
}

func (d *tcpDumper) handle(conn net.Conn) {
	defer conn.Close()
	rep := d.latest.Load()
	if rep == nil {
		return
	}
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, _ = io.WriteString(conn, dumpString(rep))
}

// Close stops serving and releases the listener. Watch calls it when its context
// is cancelled; a plain Configure never does, so the endpoint outlives the call.
func (d *tcpDumper) Close() error {
	d.stopped.Store(true)
	d.mu.Lock()
	ln := d.ln
	d.ln = nil
	d.mu.Unlock()
	if ln != nil {
		return ln.Close()
	}
	return nil
}
