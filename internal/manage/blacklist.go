package manage

import (
	"database/sql"
	"fmt"

	"cier/internal/db"
	"cier/internal/tui"
)

func BlacklistList(database *sql.DB) error {
	items, err := db.ListIgnored(database)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("Blacklist is empty")
		return nil
	}

	for _, item := range items {
		fmt.Println(item.Path)
	}

	return nil
}

func BlacklistAdd(database *sql.DB, paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("at least one path is required")
	}
	for _, path := range paths {
		if err := db.RemoveWorkflow(database, path); err != nil {
			return err
		}
		if err := db.AddIgnored(database, path); err != nil {
			return err
		}
		fmt.Printf("Added to blacklist: %s\n", path)
	}
	return nil
}

func BlacklistRestore(database *sql.DB, paths []string) error {
	if len(paths) == 0 {
		items, err := db.ListIgnored(database)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Println("Blacklist is empty")
			return nil
		}

		options := make([]string, 0, len(items))
		for _, item := range items {
			options = append(options, item.Path)
		}

		selected, err := tui.SelectIgnoredToRestore(options)
		if err != nil {
			return err
		}
		paths = selected
	}

	if len(paths) == 0 {
		fmt.Println("Nothing selected")
		return nil
	}

	for _, path := range paths {
		if err := db.RemoveIgnored(database, path); err != nil {
			return err
		}
		fmt.Printf("Removed from blacklist: %s\n", path)
	}

	return nil
}
