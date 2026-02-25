package main

import (
	"database/sql"
	"fmt"
	"os"

	"cier/internal/db"
	"cier/internal/manage"
	"cier/internal/roll"
	"cier/internal/scan"

	"github.com/spf13/cobra"
)

func main() {
	var dbPath string

	rootCmd := &cobra.Command{
		Use:   "cier",
		Short: "CIer — collect and roll out GitHub Actions workflows",
	}

	rootCmd.PersistentFlags().StringVar(
		&dbPath,
		"db",
		"",
		"SQLite database path (default: ~/.config/cier/cier.db or CIER_DB)",
	)

	rootCmd.AddCommand(scanCmd(&dbPath))
	rootCmd.AddCommand(rollCmd(&dbPath))
	rootCmd.AddCommand(removeCmd(&dbPath))
	rootCmd.AddCommand(moveCmd(&dbPath))
	rootCmd.AddCommand(blacklistCmd(&dbPath))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func openSqlite(dbPath *string) (*sql.DB, error) {
	path := *dbPath
	if path == "" {
		var err error
		path, err = db.DefaultDBPath()
		if err != nil {
			return nil, err
		}
	}
	return db.Open(path)
}

func scanCmd(dbPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan [dirs...]",
		Short: "Scan directories and add new workflows to the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openSqlite(dbPath)
			if err != nil {
				return err
			}
			defer database.Close()

			return scan.Run(database, args)
		},
	}

	return cmd
}

func rollCmd(dbPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roll",
		Short: "Collect selected workflows and open them in nvim",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openSqlite(dbPath)
			if err != nil {
				return err
			}
			defer database.Close()

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			return roll.Run(database, cwd)
		},
	}

	return cmd
}

func removeCmd(dbPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove workflows from a group and add to the blacklist",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openSqlite(dbPath)
			if err != nil {
				return err
			}
			defer database.Close()

			return manage.Remove(database)
		},
	}

	return cmd
}

func moveCmd(dbPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move workflows to another group",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openSqlite(dbPath)
			if err != nil {
				return err
			}
			defer database.Close()

			return manage.Move(database)
		},
	}

	return cmd
}

func blacklistCmd(dbPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blacklist",
		Short: "Manage the workflows blacklist",
	}

	cmd.AddCommand(blacklistListCmd(dbPath))
	cmd.AddCommand(blacklistAddCmd(dbPath))
	cmd.AddCommand(blacklistRestoreCmd(dbPath))

	return cmd
}

func blacklistListCmd(dbPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show the blacklist",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openSqlite(dbPath)
			if err != nil {
				return err
			}
			defer database.Close()

			return manage.BlacklistList(database)
		},
	}
}

func blacklistAddCmd(dbPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "add [paths...]",
		Short: "Add paths to the blacklist",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openSqlite(dbPath)
			if err != nil {
				return err
			}
			defer database.Close()

			return manage.BlacklistAdd(database, args)
		},
	}
}

func blacklistRestoreCmd(dbPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "restore [paths...]",
		Short: "Remove paths from the blacklist",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openSqlite(dbPath)
			if err != nil {
				return err
			}
			defer database.Close()

			return manage.BlacklistRestore(database, args)
		},
	}
}
