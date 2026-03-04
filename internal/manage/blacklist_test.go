package manage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cier/internal/db"
)

func TestBlacklistAddAndRestoreWithPaths(t *testing.T) {
	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	group, err := db.EnsureGroup(database, "go")
	if err != nil {
		t.Fatalf("EnsureGroup: %v", err)
	}

	wfPath := filepath.Join(tmp, "repo", ".github", "workflows", "ci.yml")
	if err := db.AddWorkflow(database, wfPath, group.ID, filepath.Join(tmp, "repo")); err != nil {
		t.Fatalf("AddWorkflow: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	rel, err := filepath.Rel(tmp, wfPath)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}

	if err := BlacklistAdd(database, []string{rel}); err != nil {
		t.Fatalf("BlacklistAdd: %v", err)
	}

	exists, err := db.WorkflowExists(database, wfPath)
	if err != nil {
		t.Fatalf("WorkflowExists: %v", err)
	}
	if exists {
		t.Fatal("workflow still exists after BlacklistAdd")
	}

	ignored, err := db.IsIgnored(database, wfPath)
	if err != nil {
		t.Fatalf("IsIgnored: %v", err)
	}
	if !ignored {
		t.Fatal("path is not ignored after BlacklistAdd")
	}

	if err := BlacklistRestore(database, []string{rel}); err != nil {
		t.Fatalf("BlacklistRestore: %v", err)
	}

	ignored, err = db.IsIgnored(database, wfPath)
	if err != nil {
		t.Fatalf("IsIgnored after restore: %v", err)
	}
	if ignored {
		t.Fatal("path is still ignored after BlacklistRestore")
	}
}

func TestBlacklistListOutput(t *testing.T) {
	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	path := filepath.Join(tmp, "x.yml")
	if err := db.AddIgnored(database, path); err != nil {
		t.Fatalf("AddIgnored: %v", err)
	}

	output := captureStdout(t, func() error {
		return BlacklistList(database)
	})

	if !strings.Contains(output, filepath.Clean(path)) {
		t.Fatalf("BlacklistList output %q does not contain %q", output, filepath.Clean(path))
	}
}

func TestBlacklistAddRequiresPaths(t *testing.T) {
	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := BlacklistAdd(database, nil); err == nil {
		t.Fatal("BlacklistAdd(nil) expected error")
	}
}

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	os.Stdout = w
	defer func() { os.Stdout = old }()

	callErr := fn()
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()

	if callErr != nil {
		t.Fatalf("captured function returned error: %v", callErr)
	}

	return buf.String()
}
