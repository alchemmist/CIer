package roll

import (
	"path/filepath"
	"testing"

	"cier/internal/db"
	"cier/internal/tui"
)

func TestCollectSelections(t *testing.T) {
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
	calls := 0
	selectGroupToEditFn = func(_ []string) (string, bool, error) {
		calls++
		if calls == 1 {
			return "go", false, nil
		}
		return "", true, nil
	}
	selectWorkflowsFn = func(_ string, _ []tui.WorkflowOption, _ []string) ([]string, error) {
		return []string{wfPath}, nil
	}
	t.Cleanup(func() {
		selectGroupToEditFn = origSelectGroup
		selectWorkflowsFn = origSelectWorkflows
	})

	selections, err := collectSelections(database, []db.Group{group})
	if err != nil {
		t.Fatalf("collectSelections: %v", err)
	}
	if len(selections) != 1 {
		t.Fatalf("selections len = %d, want 1", len(selections))
	}
	if len(selections[0].Workflows) != 1 {
		t.Fatalf("workflows len = %d, want 1", len(selections[0].Workflows))
	}
}

func TestRunUsesOpenInNvimHook(t *testing.T) {
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
	origOpenInNvim := openInNvimFn
	calls := 0
	selectGroupToEditFn = func(_ []string) (string, bool, error) {
		calls++
		if calls == 1 {
			return "go", false, nil
		}
		return "", true, nil
	}
	selectWorkflowsFn = func(_ string, _ []tui.WorkflowOption, _ []string) ([]string, error) {
		return []string{wfPath}, nil
	}
	var opened []string
	openInNvimFn = func(wf db.Workflow, targetDir string) error {
		opened = append(opened, wf.Path+"@"+targetDir)
		return nil
	}
	t.Cleanup(func() {
		selectGroupToEditFn = origSelectGroup
		selectWorkflowsFn = origSelectWorkflows
		openInNvimFn = origOpenInNvim
	})

	if err := Run(database, tmp); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(opened) != 1 {
		t.Fatalf("openInNvim calls = %d, want 1", len(opened))
	}
}
