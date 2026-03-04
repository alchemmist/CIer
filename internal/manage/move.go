package manage

import (
	"database/sql"
	"fmt"

	"cier/internal/db"
	"cier/internal/tui"
)

func Move(database *sql.DB) error {
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

	chosen, done, err := selectGroupToEditFn(groupNames)
	if err != nil {
		return err
	}
	if done {
		return nil
	}

	source := groupByName[chosen]
	workflows, err := db.ListWorkflowsByGroup(database, source.ID)
	if err != nil {
		return err
	}
	if len(workflows) == 0 {
		fmt.Printf("Group %s has no workflows\n", source.Name)
		return nil
	}

	options := make([]tui.WorkflowOption, 0, len(workflows))
	for _, wf := range workflows {
		options = append(options, tui.WorkflowOption{
			Label: fmt.Sprintf("%s — %s", wf.Path, wf.GroupName),
			Value: wf.Path,
		})
	}

	selectedPaths, err := selectWorkflowsFn(source.Name, options, nil)
	if err != nil {
		return err
	}

	if len(selectedPaths) == 0 {
		fmt.Println("Nothing selected")
		return nil
	}

	destChoice, err := selectGroupOrNewFn(groupNames)
	if err != nil {
		return err
	}

	dest, err := db.EnsureGroup(database, destChoice.Name)
	if err != nil {
		return err
	}

	for _, path := range selectedPaths {
		if err := db.MoveWorkflow(database, path, dest.ID); err != nil {
			return err
		}
		fmt.Printf("Moved: %s -> %s\n", path, dest.Name)
	}

	return nil
}
