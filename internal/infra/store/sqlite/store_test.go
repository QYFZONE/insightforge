package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"insightforge/internal/domain/session"
)

func TestOpenCreatesDatabase(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "insightforge.db")

	store, err := Open(ctx, Config{Path: path})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
}

func TestStoreCreateAndGet(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	created, err := store.Create(ctx, "  topic  ")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf("ID = %q, want %q", got.ID, created.ID)
	}
	if got.Topic != "topic" {
		t.Fatalf("Topic = %q, want %q", got.Topic, "topic")
	}
	if got.Status != session.StatusCreated {
		t.Fatalf("Status = %q, want %q", got.Status, session.StatusCreated)
	}
}

func TestStoreListOrdersByCreatedAtDesc(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	first, err := store.Create(ctx, "first")
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	time.Sleep(time.Millisecond)
	second, err := store.Create(ctx, "second")
	if err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}

	items, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].ID != second.ID || items[1].ID != first.ID {
		t.Fatalf("order = [%s, %s], want [%s, %s]", items[0].ID, items[1].ID, second.ID, first.ID)
	}
}

func TestStoreSetStatus(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	created, err := store.Create(ctx, "topic")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := store.SetStatus(ctx, created.ID, session.StatusRunning); err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status != session.StatusRunning {
		t.Fatalf("Status = %q, want %q", got.Status, session.StatusRunning)
	}
}

func TestStoreAddEventAndListEvents(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)

	created, err := store.Create(ctx, "topic")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	first, err := store.AddEvent(ctx, session.Event{
		SessionID: created.ID,
		Type:      "tool_call",
		Message:   "first",
		Payload: map[string]any{
			"tool": "search",
		},
	})
	if err != nil {
		t.Fatalf("AddEvent(first) error = %v", err)
	}
	time.Sleep(time.Millisecond)
	second, err := store.AddEvent(ctx, session.Event{
		SessionID: created.ID,
		Type:      "tool_result",
		Message:   "second",
	})
	if err != nil {
		t.Fatalf("AddEvent(second) error = %v", err)
	}

	events, err := store.ListEvents(ctx, created.ID)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].ID != first.ID || events[1].ID != second.ID {
		t.Fatalf("order = [%s, %s], want [%s, %s]", events[0].ID, events[1].ID, first.ID, second.ID)
	}
	if events[0].Payload["tool"] != "search" {
		t.Fatalf("payload tool = %v, want search", events[0].Payload["tool"])
	}
}

func TestStoreReturnsNotFound(t *testing.T) {
	store := openTestStore(t)

	_, err := store.Get(context.Background(), "missing")
	if !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Get() error = %v, want session.ErrNotFound", err)
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := Open(context.Background(), Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}
