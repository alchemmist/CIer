package scan

import (
	"path/filepath"
	"testing"

	"cier/internal/db"
)

func TestIsWorkflowFile(t *testing.T) {

	root := filepath.Join(string(filepath.Separator), "tmp", "repo")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "root workflows yml", path: filepath.Join(root, ".github", "workflows", "ci.yml"), want: true},
		{name: "nested workflows yaml", path: filepath.Join(root, "services", "api", ".github", "workflows", "deploy.yaml"), want: true},
		{name: "wrong extension", path: filepath.Join(root, ".github", "workflows", "readme.md"), want: false},
		{name: "no workflows marker", path: filepath.Join(root, "docs", "ci.yml"), want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isWorkflowFile(root, tc.path)
			if got != tc.want {
				t.Fatalf("isWorkflowFile() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRunWithNoWorkflowFiles(t *testing.T) {

	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "cier.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := Run(database, []string{tmp}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}
