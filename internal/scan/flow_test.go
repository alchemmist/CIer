package scan

import (
	"os"
	"path/filepath"
	"testing"

	"cier/internal/db"
	"cier/internal/tui"
)

func TestRunAddsWorkflowFromSelection(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	wfPath := filepath.Join(root, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(wfPath), 0o755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	if err := os.WriteFile(wfPath, []byte("name: ci\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	origSelect := selectGroupForPathFn
	origOpenReadonly := openReadonlyFn
	selectGroupForPathFn = func(_ []string, _ string, _ string) (tui.GroupChoice, error) {
		return tui.GroupChoice{Name: "go"}, nil
	}
	openReadonlyFn = func(_ string) error { return nil }
	t.Cleanup(func() {
		selectGroupForPathFn = origSelect
		openReadonlyFn = origOpenReadonly
	})

	if err := Run(database, []string{root}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	exists, err := db.WorkflowExists(database, wfPath)
	if err != nil {
		t.Fatalf("WorkflowExists: %v", err)
	}
	if !exists {
		t.Fatal("workflow was not added by scan")
	}
}

func TestRunIgnoresWorkflowWhenSelected(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	wfPath := filepath.Join(root, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(wfPath), 0o755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	if err := os.WriteFile(wfPath, []byte("name: ci\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	origSelect := selectGroupForPathFn
	origOpenReadonly := openReadonlyFn
	selectGroupForPathFn = func(_ []string, _ string, _ string) (tui.GroupChoice, error) {
		return tui.GroupChoice{Ignore: true}, nil
	}
	openReadonlyFn = func(_ string) error { return nil }
	t.Cleanup(func() {
		selectGroupForPathFn = origSelect
		openReadonlyFn = origOpenReadonly
	})

	if err := Run(database, []string{root}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	ignored, err := db.IsIgnored(database, wfPath)
	if err != nil {
		t.Fatalf("IsIgnored: %v", err)
	}
	if !ignored {
		t.Fatal("workflow was not added to blacklist")
	}

	exists, err := db.WorkflowExists(database, wfPath)
	if err != nil {
		t.Fatalf("WorkflowExists: %v", err)
	}
	if exists {
		t.Fatal("workflow exists even though it should be ignored")
	}
}

func TestRunOpenReadonlyBranch(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	wfPath := filepath.Join(root, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(wfPath), 0o755); err != nil {
		t.Fatalf("mkdir workflows: %v", err)
	}
	if err := os.WriteFile(wfPath, []byte("name: ci\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	origSelect := selectGroupForPathFn
	origOpenReadonly := openReadonlyFn
	calls := 0
	openCalled := false
	selectGroupForPathFn = func(_ []string, _ string, _ string) (tui.GroupChoice, error) {
		calls++
		if calls == 1 {
			return tui.GroupChoice{Open: true}, nil
		}
		return tui.GroupChoice{Name: "go"}, nil
	}
	openReadonlyFn = func(_ string) error {
		openCalled = true
		return nil
	}
	t.Cleanup(func() {
		selectGroupForPathFn = origSelect
		openReadonlyFn = origOpenReadonly
	})

	if err := Run(database, []string{root}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !openCalled {
		t.Fatal("openReadonly hook was not called")
	}
}
