package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Group struct {
	ID   int64
	Name string
}

type Workflow struct {
	ID          int64
	Path        string
	GroupID     int64
	GroupName   string
	ProjectRoot string
}

type Ignored struct {
	Path string
}

func DefaultDBPath() (string, error) {
	if env := os.Getenv("CIER_DB"); env != "" {
		return env, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}

	base := filepath.Join(dir, "cier")
	return filepath.Join(base, "cier.db"), nil
}

func Open(path string) (*sql.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("db path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS groups (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            created_at TEXT NOT NULL DEFAULT (datetime('now'))
        );`,
		`CREATE TABLE IF NOT EXISTS workflows (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            path TEXT NOT NULL UNIQUE,
            group_id INTEGER NOT NULL,
            project_root TEXT,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            FOREIGN KEY(group_id) REFERENCES groups(id)
        );`,
		`CREATE TABLE IF NOT EXISTS ignored (
            path TEXT NOT NULL UNIQUE,
            created_at TEXT NOT NULL DEFAULT (datetime('now'))
        );`,
		`CREATE INDEX IF NOT EXISTS idx_workflows_group_id ON workflows(group_id);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	return nil
}

func IsIgnored(db *sql.DB, path string) (bool, error) {
	var p string
	err := db.QueryRow(`SELECT path FROM ignored WHERE path = ?`, path).Scan(&p)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check ignored: %w", err)
	}
	return true, nil
}

func AddIgnored(db *sql.DB, path string) error {
	if path == "" {
		return fmt.Errorf("ignored path is empty")
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO ignored(path) VALUES (?)`, path); err != nil {
		return fmt.Errorf("insert ignored: %w", err)
	}
	return nil
}

func RemoveIgnored(db *sql.DB, path string) error {
	if _, err := db.Exec(`DELETE FROM ignored WHERE path = ?`, path); err != nil {
		return fmt.Errorf("remove ignored: %w", err)
	}
	return nil
}

func ListIgnored(db *sql.DB) ([]Ignored, error) {
	rows, err := db.Query(`SELECT path FROM ignored ORDER BY path`)
	if err != nil {
		return nil, fmt.Errorf("list ignored: %w", err)
	}
	defer rows.Close()

	var ignored []Ignored
	for rows.Next() {
		var item Ignored
		if err := rows.Scan(&item.Path); err != nil {
			return nil, fmt.Errorf("scan ignored: %w", err)
		}
		ignored = append(ignored, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list ignored rows: %w", err)
	}
	return ignored, nil
}

func ListGroups(db *sql.DB) ([]Group, error) {
	rows, err := db.Query(`SELECT id, name FROM groups ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		groups = append(groups, g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list groups rows: %w", err)
	}

	return groups, nil
}

func EnsureGroup(db *sql.DB, name string) (Group, error) {
	if name == "" {
		return Group{}, fmt.Errorf("group name is empty")
	}

	if _, err := db.Exec(`INSERT OR IGNORE INTO groups(name) VALUES (?)`, name); err != nil {
		return Group{}, fmt.Errorf("insert group: %w", err)
	}

	var g Group
	if err := db.QueryRow(`SELECT id, name FROM groups WHERE name = ?`, name).Scan(&g.ID, &g.Name); err != nil {
		return Group{}, fmt.Errorf("fetch group: %w", err)
	}

	return g, nil
}

func WorkflowExists(db *sql.DB, path string) (bool, error) {
	var id int64
	err := db.QueryRow(`SELECT id FROM workflows WHERE path = ?`, path).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check workflow: %w", err)
	}
	return true, nil
}

func AddWorkflow(db *sql.DB, path string, groupID int64, projectRoot string) error {
	if path == "" {
		return fmt.Errorf("workflow path is empty")
	}

	if _, err := db.Exec(`INSERT OR IGNORE INTO workflows(path, group_id, project_root) VALUES (?, ?, ?)`, path, groupID, projectRoot); err != nil {
		return fmt.Errorf("insert workflow: %w", err)
	}
	return nil
}

func ListWorkflowsByGroup(db *sql.DB, groupID int64) ([]Workflow, error) {
	rows, err := db.Query(`
        SELECT w.id, w.path, w.group_id, g.name, w.project_root
        FROM workflows w
        JOIN groups g ON g.id = w.group_id
        WHERE w.group_id = ?
        ORDER BY w.path
    `, groupID)
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	defer rows.Close()

	var workflows []Workflow
	for rows.Next() {
		var w Workflow
		if err := rows.Scan(&w.ID, &w.Path, &w.GroupID, &w.GroupName, &w.ProjectRoot); err != nil {
			return nil, fmt.Errorf("scan workflow: %w", err)
		}
		workflows = append(workflows, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list workflows rows: %w", err)
	}

	return workflows, nil
}

func ListAllWorkflows(db *sql.DB) ([]Workflow, error) {
	rows, err := db.Query(`
        SELECT w.id, w.path, w.group_id, g.name, w.project_root
        FROM workflows w
        JOIN groups g ON g.id = w.group_id
        ORDER BY g.name COLLATE NOCASE, w.path
    `)
	if err != nil {
		return nil, fmt.Errorf("list all workflows: %w", err)
	}
	defer rows.Close()

	var workflows []Workflow
	for rows.Next() {
		var w Workflow
		if err := rows.Scan(&w.ID, &w.Path, &w.GroupID, &w.GroupName, &w.ProjectRoot); err != nil {
			return nil, fmt.Errorf("scan workflow: %w", err)
		}
		workflows = append(workflows, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list workflows rows: %w", err)
	}

	return workflows, nil
}

func RemoveWorkflow(db *sql.DB, path string) error {
	if _, err := db.Exec(`DELETE FROM workflows WHERE path = ?`, path); err != nil {
		return fmt.Errorf("remove workflow: %w", err)
	}
	return nil
}

func MoveWorkflow(db *sql.DB, path string, groupID int64) error {
	if _, err := db.Exec(`UPDATE workflows SET group_id = ? WHERE path = ?`, groupID, path); err != nil {
		return fmt.Errorf("move workflow: %w", err)
	}
	return nil
}
