package vault

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/goreflect/gostructor"
)

// mutableLogical is a fake Vault whose secret data can change at runtime.
type mutableLogical struct {
	mu      sync.Mutex
	secrets map[string]*vaultapi.Secret
}

func (m *mutableLogical) Read(path string) (*vaultapi.Secret, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.secrets[path], nil
}

func (m *mutableLogical) set(path string, data map[string]any) {
	m.mu.Lock()
	m.secrets[path] = &vaultapi.Secret{Data: data}
	m.mu.Unlock()
}

func TestSourceWatchDetectsSecretChange(t *testing.T) {
	fake := &mutableLogical{secrets: map[string]*vaultapi.Secret{
		"my-service/stage": {Data: map[string]any{"api-key": "old"}},
	}}
	s := &source{
		newClient: func() (logicalReader, error) { return fake, nil },
		poll:      15 * time.Millisecond,
		log:       slog.New(slog.DiscardHandler),
	}

	// Initial resolution records the referenced path so Watch knows to poll it.
	field := fieldWithVaultTag("my-service/stage#api-key", reflect.TypeOf(""))
	if _, found, err := s.Resolve(field); err != nil || !found {
		t.Fatalf("initial resolve: found=%v err=%v", found, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 4)
	go s.Watch(ctx, func() { changed <- struct{}{} })

	// Let PollWatch establish its baseline, then rotate the secret.
	time.Sleep(60 * time.Millisecond)
	fake.set("my-service/stage", map[string]any{"api-key": "rotated"})

	select {
	case <-changed:
	case <-time.After(3 * time.Second):
		t.Fatal("Watch did not detect the secret change")
	}
}

func TestSourceWatchIsWatchable(t *testing.T) {
	// New must return something that satisfies gostructor.Watchable, so the
	// Watch driver picks it up.
	if _, ok := New().(gostructor.Watchable); !ok {
		t.Fatal("vault.New() does not implement gostructor.Watchable")
	}
}
