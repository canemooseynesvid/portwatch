package tag_test

import (
	"strings"
	"testing"

	"portwatch/internal/alerting"
	"portwatch/internal/tag"
)

func TestStore_SetAndGet(t *testing.T) {
	s := tag.New()
	s.Set("tcp/80", "env", "prod")
	s.Set("tcp/80", "team", "infra")

	tags, ok := s.Get("tcp/80")
	if !ok {
		t.Fatal("expected tags to exist")
	}
	if tags["env"] != "prod" {
		t.Errorf("env: got %q, want %q", tags["env"], "prod")
	}
	if tags["team"] != "infra" {
		t.Errorf("team: got %q, want %q", tags["team"], "infra")
	}
}

func TestStore_GetMissing(t *testing.T) {
	s := tag.New()
	_, ok := s.Get("udp/53")
	if ok {
		t.Error("expected no tags for unknown entity")
	}
}

func TestStore_Delete(t *testing.T) {
	s := tag.New()
	s.Set("tcp/443", "tls", "true")
	s.Delete("tcp/443")
	_, ok := s.Get("tcp/443")
	if ok {
		t.Error("expected tags to be deleted")
	}
}

func TestStore_Len(t *testing.T) {
	s := tag.New()
	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
	s.Set("tcp/80", "k", "v")
	s.Set("udp/53", "k", "v")
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
}

func TestTags_String(t *testing.T) {
	tags := tag.Tags{"env": "prod"}
	s := tags.String()
	if !strings.Contains(s, "env=prod") {
		t.Errorf("unexpected string: %s", s)
	}
}

func TestTags_String_Empty(t *testing.T) {
	tags := tag.Tags{}
	if tags.String() != "{}" {
		t.Errorf("unexpected string: %s", tags.String())
	}
}

func TestMiddleware_InjectsTags(t *testing.T) {
	store := tag.New()
	store.Set("tcp/8080", "service", "api")

	var got alerting.Alert
	next := alerting.HandlerFunc(func(a alerting.Alert) error {
		got = a
		return nil
	})

	mw := tag.NewMiddleware(store, next)
	a := alerting.Alert{
		Message: "port bound",
		Meta:    map[string]string{"port": "8080", "protocol": "tcp"},
		Tags:    map[string]string{},
	}
	if err := mw.Handle(a); err != nil {
		t.Fatal(err)
	}
	if got.Tags["service"] != "api" {
		t.Errorf("service tag: got %q, want %q", got.Tags["service"], "api")
	}
}

func TestMiddleware_NoTagsForUnknownEntity(t *testing.T) {
	store := tag.New()

	var got alerting.Alert
	next := alerting.HandlerFunc(func(a alerting.Alert) error {
		got = a
		return nil
	})

	mw := tag.NewMiddleware(store, next)
	a := alerting.Alert{
		Message: "port bound",
		Meta:    map[string]string{"port": "9999", "protocol": "tcp"},
		Tags:    map[string]string{"existing": "yes"},
	}
	if err := mw.Handle(a); err != nil {
		t.Fatal(err)
	}
	if len(got.Tags) != 1 || got.Tags["existing"] != "yes" {
		t.Errorf("unexpected tags: %v", got.Tags)
	}
}
