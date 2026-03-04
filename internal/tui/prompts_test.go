package tui

import (
	"testing"

	"github.com/charmbracelet/huh"
)

func TestGroupOptionsContent(t *testing.T) {
	groups := []string{"backend", "frontend", "ops"}
	options := groupOptions(groups)

	if len(options) != len(groups) {
		t.Fatalf("groupOptions len = %d, want %d", len(options), len(groups))
	}
	for i, group := range groups {
		if options[i].Value != group {
			t.Fatalf("options[%d].Value = %q, want %q", i, options[i].Value, group)
		}
		wantKey := "· " + group
		if options[i].Key != wantKey {
			t.Fatalf("options[%d].Key = %q, want %q", i, options[i].Key, wantKey)
		}
	}
}

func TestVimKeyMapConfigured(t *testing.T) {
	km := vimKeyMap()
	if km == nil {
		t.Fatal("vimKeyMap returned nil")
	}
	if len(km.Select.Up.Keys()) == 0 {
		t.Fatal("Select.Up has no keys")
	}
	if len(km.Select.Down.Keys()) == 0 {
		t.Fatal("Select.Down has no keys")
	}
	if len(km.MultiSelect.Up.Keys()) == 0 {
		t.Fatal("MultiSelect.Up has no keys")
	}
	if len(km.MultiSelect.Down.Keys()) == 0 {
		t.Fatal("MultiSelect.Down has no keys")
	}
}

func TestNewForm(t *testing.T) {
	form := newForm(
		huh.NewGroup(
			huh.NewInput().Title("name"),
		),
	)
	if form == nil {
		t.Fatal("newForm returned nil")
	}
}
