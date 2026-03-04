package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultDBPathUsesEnv(t *testing.T) {
	t.Setenv("CIER_DB", "/tmp/custom-cier.db")
	path, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath returned error: %v", err)
	}
	if path != "/tmp/custom-cier.db" {
		t.Fatalf("DefaultDBPath = %q, want %q", path, "/tmp/custom-cier.db")
	}
}

func TestDefaultDBPathFallback(t *testing.T) {
	t.Setenv("CIER_DB", "")
	path, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("DefaultDBPath returned error: %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join("cier", "cier.db")) {
		t.Fatalf("DefaultDBPath = %q, want suffix %q", path, filepath.Join("cier", "cier.db"))
	}
}

func TestOpenEmptyPath(t *testing.T) {
	_, err := Open("")
	if err == nil {
		t.Fatal("Open(\"\") expected error, got nil")
	}
}

func TestDBWorkflowAndBlacklistLifecycle(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "cier.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	alpha, err := EnsureGroup(database, "alpha")
	if err != nil {
		t.Fatalf("EnsureGroup alpha: %v", err)
	}
	beta, err := EnsureGroup(database, "beta")
	if err != nil {
		t.Fatalf("EnsureGroup beta: %v", err)
	}

	groups, err := ListGroups(database)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("ListGroups len = %d, want 2", len(groups))
	}

	repoRoot := filepath.Join(tmp, "repo")
	wfAbs := filepath.Join(repoRoot, ".github", "workflows", "ci.yml")

	if err := AddWorkflow(database, wfAbs, alpha.ID, repoRoot); err != nil {
		t.Fatalf("AddWorkflow: %v", err)
	}

	exists, err := WorkflowExists(database, wfAbs)
	if err != nil {
		t.Fatalf("WorkflowExists(abs): %v", err)
	}
	if !exists {
		t.Fatal("WorkflowExists(abs) = false, want true")
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	wfRel, err := filepath.Rel(tmp, wfAbs)
	if err != nil {
		t.Fatalf("filepath.Rel: %v", err)
	}

	exists, err = WorkflowExists(database, wfRel)
	if err != nil {
		t.Fatalf("WorkflowExists(rel): %v", err)
	}
	if !exists {
		t.Fatal("WorkflowExists(rel) = false, want true")
	}

	if err := MoveWorkflow(database, wfRel, beta.ID); err != nil {
		t.Fatalf("MoveWorkflow(rel): %v", err)
	}

	betaWorkflows, err := ListWorkflowsByGroup(database, beta.ID)
	if err != nil {
		t.Fatalf("ListWorkflowsByGroup(beta): %v", err)
	}
	if len(betaWorkflows) != 1 {
		t.Fatalf("beta workflows len = %d, want 1", len(betaWorkflows))
	}
	if betaWorkflows[0].GroupName != "beta" {
		t.Fatalf("group name = %q, want beta", betaWorkflows[0].GroupName)
	}
	if betaWorkflows[0].Path != filepath.Clean(wfAbs) {
		t.Fatalf("stored path = %q, want %q", betaWorkflows[0].Path, filepath.Clean(wfAbs))
	}

	all, err := ListAllWorkflows(database)
	if err != nil {
		t.Fatalf("ListAllWorkflows: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("ListAllWorkflows len = %d, want 1", len(all))
	}

	if err := AddIgnored(database, wfAbs); err != nil {
		t.Fatalf("AddIgnored(abs): %v", err)
	}

	ignored, err := IsIgnored(database, wfRel)
	if err != nil {
		t.Fatalf("IsIgnored(rel): %v", err)
	}
	if !ignored {
		t.Fatal("IsIgnored(rel) = false, want true")
	}

	ignoredItems, err := ListIgnored(database)
	if err != nil {
		t.Fatalf("ListIgnored: %v", err)
	}
	if len(ignoredItems) != 1 {
		t.Fatalf("ListIgnored len = %d, want 1", len(ignoredItems))
	}
	if ignoredItems[0].Path != filepath.Clean(wfAbs) {
		t.Fatalf("ignored path = %q, want %q", ignoredItems[0].Path, filepath.Clean(wfAbs))
	}

	if err := RemoveIgnored(database, wfRel); err != nil {
		t.Fatalf("RemoveIgnored(rel): %v", err)
	}

	ignored, err = IsIgnored(database, wfAbs)
	if err != nil {
		t.Fatalf("IsIgnored(abs) after remove: %v", err)
	}
	if ignored {
		t.Fatal("IsIgnored(abs) = true after remove")
	}

	if err := RemoveWorkflow(database, wfRel); err != nil {
		t.Fatalf("RemoveWorkflow(rel): %v", err)
	}

	exists, err = WorkflowExists(database, wfAbs)
	if err != nil {
		t.Fatalf("WorkflowExists(abs) after remove: %v", err)
	}
	if exists {
		t.Fatal("WorkflowExists(abs) = true after remove")
	}
}

func TestValidationErrors(t *testing.T) {
	tmp := t.TempDir()
	database, err := Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if _, err := EnsureGroup(database, ""); err == nil {
		t.Fatal("EnsureGroup empty name expected error")
	}
	if err := AddIgnored(database, ""); err == nil {
		t.Fatal("AddIgnored empty path expected error")
	}
	if err := AddWorkflow(database, "", 1, ""); err == nil {
		t.Fatal("AddWorkflow empty path expected error")
	}
	if err := RemoveIgnored(database, ""); err == nil {
		t.Fatal("RemoveIgnored empty path expected error")
	}
	if err := RemoveWorkflow(database, ""); err == nil {
		t.Fatal("RemoveWorkflow empty path expected error")
	}
	if err := MoveWorkflow(database, "", 1); err == nil {
		t.Fatal("MoveWorkflow empty path expected error")
	}
	if _, err := WorkflowExists(database, ""); err == nil {
		t.Fatal("WorkflowExists empty path expected error")
	}
	if _, err := IsIgnored(database, ""); err == nil {
		t.Fatal("IsIgnored empty path expected error")
	}
}
