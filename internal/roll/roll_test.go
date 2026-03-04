package roll

import (
	"path/filepath"
	"testing"

	"cier/internal/db"
)

func TestFlattenSelections(t *testing.T) {

	in := []Selection{
		{Workflows: []db.Workflow{{Path: "a"}, {Path: "b"}}},
		{Workflows: []db.Workflow{{Path: "c"}}},
	}

	got := flattenSelections(in)
	if len(got) != 3 {
		t.Fatalf("flattenSelections len = %d, want 3", len(got))
	}
	if got[0].Path != "a" || got[1].Path != "b" || got[2].Path != "c" {
		t.Fatalf("flattenSelections order mismatch: %#v", got)
	}
}

func TestWorkflowLabel(t *testing.T) {

	wf := db.Workflow{
		Path:        filepath.Join("/tmp", "repo", ".github", "workflows", "ci.yml"),
		ProjectRoot: filepath.Join("/tmp", "repo"),
	}
	if got := workflowLabel(wf); got != "ci.yml — repo" {
		t.Fatalf("workflowLabel() = %q, want %q", got, "ci.yml — repo")
	}

	wf.ProjectRoot = ""
	if got := workflowLabel(wf); got != "ci.yml" {
		t.Fatalf("workflowLabel() without project = %q, want %q", got, "ci.yml")
	}
}

func TestQuoteLines(t *testing.T) {

	got := quoteLines([]string{"plain", "it' s", "x'y"})
	want := []string{"'plain'", "'it'' s'", "'x''y'"}
	if len(got) != len(want) {
		t.Fatalf("quoteLines len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("quoteLines[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestVimReadCmdEscapesQuotes(t *testing.T) {

	got := vimReadCmd("/tmp/wo'rkflow.yml")
	want := "execute '0r ' .. fnameescape('/tmp/wo''rkflow.yml')"
	if got != want {
		t.Fatalf("vimReadCmd() = %q, want %q", got, want)
	}
}
