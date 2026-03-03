package scan

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cier/internal/db"
	"cier/internal/tui"
)

func Run(database *sql.DB, roots []string) error {
	if len(roots) == 0 {
		roots = []string{"."}
	}

	for _, root := range roots {
		if err := scanRoot(database, root); err != nil {
			if err == errStopScan {
				return nil
			}
			return err
		}
	}

	return nil
}

func scanRoot(database *sql.DB, root string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve root %q: %w", root, err)
	}
	root = filepath.Clean(absRoot)

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !isWorkflowFile(root, path) {
			return nil
		}

		ignored, err := db.IsIgnored(database, path)
		if err != nil {
			return err
		}
		if ignored {
			return nil
		}

		exists, err := db.WorkflowExists(database, path)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}

		groups, err := db.ListGroups(database)
		if err != nil {
			return err
		}

		groupNames := make([]string, 0, len(groups))
		for _, g := range groups {
			groupNames = append(groupNames, g.Name)
		}

		for {
			choice, err := tui.SelectGroupForPath(groupNames, path, root)
			if err != nil {
				return err
			}

			if choice.Open {
				if err := openReadonly(path); err != nil {
					return err
				}
				continue
			}

			if choice.Stop {
				return errStopScan
			}

			if choice.Ignore {
				if err := db.AddIgnored(database, path); err != nil {
					return err
				}
				fmt.Printf("Added to blacklist: %s\n", path)
				return nil
			}

			group, err := db.EnsureGroup(database, choice.Name)
			if err != nil {
				return err
			}

			if err := db.AddWorkflow(database, path, group.ID, root); err != nil {
				return err
			}

			fmt.Printf("Added workflow: %s -> group %s\n", path, group.Name)
			return nil
		}
	})
}

var errStopScan = fmt.Errorf("stop scan")

func openReadonly(path string) error {
	cmd := exec.Command("nvim", "-R", "-c", "setlocal nomodifiable", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("open nvim: %w", err)
	}
	return nil
}

func isWorkflowFile(root, path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yml" && ext != ".yaml" {
		return false
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	sep := string(filepath.Separator)
	marker := sep + ".github" + sep + "workflows" + sep

	if strings.HasPrefix(rel, ".github"+sep+"workflows"+sep) {
		return true
	}
	return strings.Contains(rel, marker)
}
