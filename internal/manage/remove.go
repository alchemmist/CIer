package manage

import (
	"database/sql"
	"fmt"

	"cier/internal/db"
	"cier/internal/tui"
)

func Remove(database *sql.DB) error {
	groups, err := db.ListGroups(database)
	if err != nil {
		return err
	}
	if len(groups) == 0 {
		return fmt.Errorf("no groups in the database")
	}

	groupNames := make([]string, 0, len(groups))
	groupByName := map[string]db.Group{}
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
		groupByName[g.Name] = g
	}

	chosen, done, err := tui.SelectGroupToEdit(groupNames)
	if err != nil {
		return err
	}
	if done {
		return nil
	}

	group := groupByName[chosen]
	workflows, err := db.ListWorkflowsByGroup(database, group.ID)
	if err != nil {
		return err
	}
	if len(workflows) == 0 {
		fmt.Printf("Group %s has no workflows\n", group.Name)
		return nil
	}

	options := make([]tui.WorkflowOption, 0, len(workflows))
	for _, wf := range workflows {
		options = append(options, tui.WorkflowOption{
			Label: fmt.Sprintf("%s — %s", wf.Path, wf.GroupName),
			Value: wf.Path,
		})
	}

	selectedPaths, err := tui.SelectWorkflows(group.Name, options, nil)
	if err != nil {
		return err
	}

	if len(selectedPaths) == 0 {
		fmt.Println("Nothing selected")
		return nil
	}

	for _, path := range selectedPaths {
		if err := db.RemoveWorkflow(database, path); err != nil {
			return err
		}
		if err := db.AddIgnored(database, path); err != nil {
			return err
		}
		fmt.Printf("Removed and added to blacklist: %s\n", path)
	}

	return nil
}
