package main

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestOpenSqliteWithExplicitPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	dbPath := path

	database, err := openSqlite(&dbPath)
	if err != nil {
		t.Fatalf("openSqlite returned error: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
}

func TestOpenSqliteUsesEnvDefaultPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "from-env.db")
	t.Setenv("CIER_DB", path)
	dbPath := ""

	database, err := openSqlite(&dbPath)
	if err != nil {
		t.Fatalf("openSqlite returned error: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
}

func TestBlacklistAddCmdRequiresArgs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	cmd := blacklistAddCmd(&dbPath)
	if cmd.Args == nil {
		t.Fatal("blacklist add command args validator is nil")
	}
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Fatal("expected args validation error for no paths")
	}
}

func TestCommandMetadata(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	cases := []*cobra.Command{
		scanCmd(&dbPath),
		rollCmd(&dbPath),
		removeCmd(&dbPath),
		moveCmd(&dbPath),
		blacklistCmd(&dbPath),
		blacklistListCmd(&dbPath),
		blacklistAddCmd(&dbPath),
		blacklistRestoreCmd(&dbPath),
	}

	for _, cmd := range cases {
		if cmd.Use == "" {
			t.Fatalf("command has empty Use: %+v", cmd)
		}
		if cmd.Short == "" {
			t.Fatalf("command has empty Short: %s", cmd.Use)
		}
	}
}
