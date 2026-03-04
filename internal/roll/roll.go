package roll

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cier/internal/db"
	"cier/internal/tui"
)

var (
	selectGroupToEditFn = tui.SelectGroupToEdit
	selectWorkflowsFn   = tui.SelectWorkflows
	openInNvimFn        = openInNvim
)

type Selection struct {
	Group     db.Group
	Workflows []db.Workflow
}

func Run(database *sql.DB, destRoot string) error {
	groups, err := db.ListGroups(database)
	if err != nil {
		return err
	}
	if len(groups) == 0 {
		return fmt.Errorf("no groups in the database, run scan first")
	}

	selections, err := collectSelections(database, groups)
	if err != nil {
		return err
	}
	if len(selections) == 0 {
		fmt.Println("Nothing selected")
		return nil
	}

	workflows := flattenSelections(selections)
	if len(workflows) == 0 {
		fmt.Println("Nothing selected")
		return nil
	}

	targetDir := filepath.Join(destRoot, ".github", "workflows")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create workflows dir: %w", err)
	}

	for _, wf := range workflows {
		if err := openInNvimFn(wf, targetDir); err != nil {
			return err
		}
	}

	return nil
}

func collectSelections(database *sql.DB, groups []db.Group) ([]Selection, error) {
	selectedByGroup := map[int64][]string{}
	var result []Selection

	groupNames := make([]string, 0, len(groups))
	groupByName := map[string]db.Group{}
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
		groupByName[g.Name] = g
	}

	for {
		chosen, done, err := selectGroupToEditFn(groupNames)
		if err != nil {
			return nil, err
		}
		if done {
			break
		}

		group := groupByName[chosen]
		workflows, err := db.ListWorkflowsByGroup(database, group.ID)
		if err != nil {
			return nil, err
		}
		if len(workflows) == 0 {
			fmt.Printf("Group %s has no workflows\n", group.Name)
			continue
		}

		options := make([]tui.WorkflowOption, 0, len(workflows))
		for _, wf := range workflows {
			options = append(options, tui.WorkflowOption{
				Label: workflowLabel(wf),
				Value: wf.Path,
			})
		}

		selectedPaths, err := selectWorkflowsFn(group.Name, options, selectedByGroup[group.ID])
		if err != nil {
			return nil, err
		}

		selectedByGroup[group.ID] = selectedPaths
	}

	for _, g := range groups {
		paths := selectedByGroup[g.ID]
		if len(paths) == 0 {
			continue
		}

		workflows, err := db.ListWorkflowsByGroup(database, g.ID)
		if err != nil {
			return nil, err
		}

		lookup := map[string]db.Workflow{}
		for _, wf := range workflows {
			lookup[wf.Path] = wf
		}

		selection := Selection{Group: g}
		for _, path := range paths {
			if wf, ok := lookup[path]; ok {
				selection.Workflows = append(selection.Workflows, wf)
			}
		}

		if len(selection.Workflows) > 0 {
			result = append(result, selection)
		}
	}

	return result, nil
}

func flattenSelections(selections []Selection) []db.Workflow {
	var workflows []db.Workflow
	for _, sel := range selections {
		workflows = append(workflows, sel.Workflows...)
	}
	return workflows
}

func workflowLabel(wf db.Workflow) string {
	name := filepath.Base(wf.Path)
	project := filepath.Base(wf.ProjectRoot)
	if project == "." || project == string(filepath.Separator) || project == "" {
		project = wf.ProjectRoot
	}
	if project == "" {
		return name
	}
	return fmt.Sprintf("%s — %s", name, project)
}

func openInNvim(wf db.Workflow, targetDir string) error {
	header := []string{
		fmt.Sprintf("# CIer: group %q", wf.GroupName),
	}
	if wf.ProjectRoot != "" {
		header = append(header, fmt.Sprintf("# CIer: project %q", wf.ProjectRoot))
	}
	header = append(header, fmt.Sprintf("# CIer: source %q", wf.Path))
	header = append(header, "")

	headerCmd := fmt.Sprintf("call append(0, [%s])", strings.Join(quoteLines(header), ","))

	cmd := exec.Command("nvim",
		"-c", "enew",
		"-c", vimReadCmd(wf.Path),
		"-c", headerCmd,
	)
	if targetDir != "" {
		cmd.Dir = targetDir
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("open nvim: %w", err)
	}

	return nil
}

func quoteLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, fmt.Sprintf("'%s'", strings.ReplaceAll(line, "'", "''")))
	}
	return out
}

func vimReadCmd(path string) string {
	return fmt.Sprintf("execute '0r ' .. fnameescape('%s')", strings.ReplaceAll(path, "'", "''"))
}
