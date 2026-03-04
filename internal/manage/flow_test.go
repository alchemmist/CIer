package manage

import (
	"path/filepath"
	"testing"

	"cier/internal/db"
	"cier/internal/tui"
)

func TestRemoveMovesWorkflowToBlacklist(t *testing.T) {
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

	origSelectGroup := selectGroupToEditFn
	origSelectWorkflows := selectWorkflowsFn
	selectGroupToEditFn = func(_ []string) (string, bool, error) { return "go", false, nil }
	selectWorkflowsFn = func(_ string, _ []tui.WorkflowOption, _ []string) ([]string, error) {
		return []string{wfPath}, nil
	}
	t.Cleanup(func() {
		selectGroupToEditFn = origSelectGroup
		selectWorkflowsFn = origSelectWorkflows
	})

	if err := Remove(database); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	exists, err := db.WorkflowExists(database, wfPath)
	if err != nil {
		t.Fatalf("WorkflowExists: %v", err)
	}
	if exists {
		t.Fatal("workflow still exists after Remove")
	}
	ignored, err := db.IsIgnored(database, wfPath)
	if err != nil {
		t.Fatalf("IsIgnored: %v", err)
	}
	if !ignored {
		t.Fatal("path was not added to blacklist")
	}
}

func TestMoveTransfersWorkflowBetweenGroups(t *testing.T) {
	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	src, err := db.EnsureGroup(database, "src")
	if err != nil {
		t.Fatalf("EnsureGroup(src): %v", err)
	}
	dst, err := db.EnsureGroup(database, "dst")
	if err != nil {
		t.Fatalf("EnsureGroup(dst): %v", err)
	}
	wfPath := filepath.Join(tmp, "repo", ".github", "workflows", "ci.yml")
	if err := db.AddWorkflow(database, wfPath, src.ID, filepath.Join(tmp, "repo")); err != nil {
		t.Fatalf("AddWorkflow: %v", err)
	}

	origSelectGroup := selectGroupToEditFn
	origSelectWorkflows := selectWorkflowsFn
	origSelectGroupOrNew := selectGroupOrNewFn
	selectGroupToEditFn = func(_ []string) (string, bool, error) { return "src", false, nil }
	selectWorkflowsFn = func(_ string, _ []tui.WorkflowOption, _ []string) ([]string, error) {
		return []string{wfPath}, nil
	}
	selectGroupOrNewFn = func(_ []string) (tui.GroupChoice, error) {
		return tui.GroupChoice{Name: "dst"}, nil
	}
	t.Cleanup(func() {
		selectGroupToEditFn = origSelectGroup
		selectWorkflowsFn = origSelectWorkflows
		selectGroupOrNewFn = origSelectGroupOrNew
	})

	if err := Move(database); err != nil {
		t.Fatalf("Move: %v", err)
	}

	srcWorkflows, err := db.ListWorkflowsByGroup(database, src.ID)
	if err != nil {
		t.Fatalf("ListWorkflowsByGroup(src): %v", err)
	}
	if len(srcWorkflows) != 0 {
		t.Fatalf("src workflows len = %d, want 0", len(srcWorkflows))
	}

	dstWorkflows, err := db.ListWorkflowsByGroup(database, dst.ID)
	if err != nil {
		t.Fatalf("ListWorkflowsByGroup(dst): %v", err)
	}
	if len(dstWorkflows) != 1 {
		t.Fatalf("dst workflows len = %d, want 1", len(dstWorkflows))
	}
}

func TestBlacklistRestoreUsesSelectorWhenNoPathsProvided(t *testing.T) {
	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	path := filepath.Join(tmp, "repo", ".github", "workflows", "ci.yml")
	if err := db.AddIgnored(database, path); err != nil {
		t.Fatalf("AddIgnored: %v", err)
	}

	origSelectIgnored := selectIgnoredToRestFn
	selectIgnoredToRestFn = func(_ []string) ([]string, error) {
		return []string{path}, nil
	}
	t.Cleanup(func() { selectIgnoredToRestFn = origSelectIgnored })

	if err := BlacklistRestore(database, nil); err != nil {
		t.Fatalf("BlacklistRestore: %v", err)
	}

	ignored, err := db.IsIgnored(database, path)
	if err != nil {
		t.Fatalf("IsIgnored: %v", err)
	}
	if ignored {
		t.Fatal("path is still ignored after restore")
	}
}
