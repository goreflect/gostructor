package gostructor

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// Watch fills target once, then keeps a long-lived service's configuration live:
// whenever any Watchable source among the configured sources reports a change,
// it re-fills a *fresh* copy of T through the exact same resolution, hook,
// trace, and masking path as the initial fill, and delivers the result to
// onReload. It blocks until ctx is cancelled, then returns ctx.Err().
//
// The reload is transactional (last-known-good):
//
//  1. Resolve into a fresh *T — the live struct is never mutated in place, so a
//     half-applied fill is never observable.
//  2. Run WithValidate (whole-struct) against that fresh copy.
//  3. Only on success is it published via onReload(fresh, nil). On failure the
//     previously good config keeps serving and the error is delivered via
//     onReload(nil, err) (and logged if WithLogger is set) — a bad edit is a
//     non-fatal, logged event, not a crash.
//
// onReload is called once with the initial fill (so a caller can publish it
// with the same atomic-swap code it uses for reloads) and again for every
// reload attempt thereafter. A typical caller stores the latest good *T in an
// atomic.Pointer and swaps it on each onReload(cfg, nil).
//
// If the *initial* fill fails, Watch returns that error immediately without
// calling onReload — startup misconfiguration stays fatal, just like Configure.
//
// WithDebounce coalesces a burst of change signals into a single reload;
// without it every signal reloads immediately. If no configured source is
// Watchable, Watch performs the initial fill, delivers it, and then blocks
// until ctx is cancelled (there is nothing to watch).
func Watch[T any](ctx context.Context, target *T, onReload func(*T, error), opts ...Option) error {
	cfg := newConfig(opts)

	// Initial fill into the caller's target; a startup failure is fatal.
	if _, _, err := runConfigure(cfg, target, false); err != nil {
		return err
	}
	if cfg.validate != nil {
		if err := cfg.validate(target); err != nil {
			return err
		}
	}
	deliver(onReload, target, nil)

	// Fan every Watchable source's change signals into one channel. A buffer of
	// one plus the non-blocking send collapses simultaneous signals; the
	// debounce window below does the time-based coalescing.
	changes := make(chan struct{}, 1)
	notify := func() {
		select {
		case changes <- struct{}{}:
		default:
		}
	}

	watched := 0
	for _, s := range cfg.sources {
		w, ok := s.(Watchable)
		if !ok {
			continue
		}
		watched++
		go func(w Watchable) {
			if err := w.Watch(ctx, notify); err != nil && !errors.Is(err, context.Canceled) {
				cfg.logger.Error("gostructor: watch source failed", "source", sourceName(w), "err", err)
			}
		}(w)
	}
	cfg.logger.Debug("gostructor: watching for live config changes", "watchable_sources", watched)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changes:
			if !waitQuiet(ctx, changes, cfg.debounce) {
				return ctx.Err()
			}
			reloadOnce(cfg, onReload)
		}
	}
}

// reloadOnce performs one transactional reload: fill a fresh *T, validate it,
// and publish it via onReload — or, on failure, deliver the error and leave the
// previously published config serving.
func reloadOnce[T any](cfg *config, onReload func(*T, error)) {
	fresh := new(T)
	if _, _, err := runConfigure(cfg, fresh, false); err != nil {
		cfg.logger.Error("gostructor: reload failed, keeping last-known-good", "err", err)
		deliver(onReload, nil, err)
		return
	}
	if cfg.validate != nil {
		if err := cfg.validate(fresh); err != nil {
			cfg.logger.Error("gostructor: reload rejected by validation, keeping last-known-good", "err", err)
			deliver(onReload, (*T)(nil), err)
			return
		}
	}
	cfg.logger.Debug("gostructor: config reloaded")
	deliver(onReload, fresh, nil)
}

// waitQuiet blocks until the change signal has been quiet for the debounce
// window, restarting the window on every fresh signal so a burst collapses into
// one reload. It returns false if ctx is cancelled while waiting. With a
// non-positive window it returns true immediately (no debouncing).
func waitQuiet(ctx context.Context, changes <-chan struct{}, window time.Duration) bool {
	if window <= 0 {
		return true
	}
	timer := time.NewTimer(window)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return false
		case <-changes:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(window)
		case <-timer.C:
			return true
		}
	}
}

// deliver calls onReload if the caller supplied one.
func deliver[T any](onReload func(*T, error), cfg *T, err error) {
	if onReload != nil {
		onReload(cfg, err)
	}
}

// sourceName reports a Watchable's source name for logs when it is also a
// Source, else a placeholder.
func sourceName(w Watchable) string {
	if s, ok := w.(Source); ok {
		return s.Name()
	}
	return "unknown"
}

// PollWatch is a helper a Watchable source can use to implement Watch by
// polling a cheap fingerprint (a commit SHA, a KV modify-index, a secret
// version) on an interval and calling onChange whenever it differs from the
// previous poll. It blocks until ctx is cancelled and returns ctx.Err(); a
// fingerprint error is logged (if log is non-nil) and retried on the next tick
// rather than ending the watch, so a transient upstream outage doesn't stop
// live reloads once it recovers. A non-positive interval defaults to 30s.
//
// The first successful fingerprint establishes the baseline and does not fire
// onChange (the initial value is already reflected in the first fill).
func PollWatch(ctx context.Context, interval time.Duration, fingerprint func(context.Context) (string, error), onChange func(), log *slog.Logger) error {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var last string
	haveBaseline := false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			fp, err := fingerprint(ctx)
			if err != nil {
				if log != nil {
					log.Warn("gostructor: poll fingerprint failed", "err", err)
				}
				continue
			}
			if !haveBaseline {
				last, haveBaseline = fp, true
				continue
			}
			if fp != last {
				last = fp
				onChange()
			}
		}
	}
}
