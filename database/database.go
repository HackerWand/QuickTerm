package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"QuickTerm/models"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, ".quickterm.db")
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			ignored_commands TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS command_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			workspace_id INTEGER NOT NULL,
			UNIQUE(name, workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			content TEXT NOT NULL,
			description TEXT,
			group_id INTEGER,
			workspace_id INTEGER NOT NULL,
			template_params TEXT DEFAULT '',
			UNIQUE(content, workspace_id),
			FOREIGN KEY (group_id) REFERENCES command_groups(id) ON DELETE SET NULL,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS recent_paths (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workspace_id INTEGER NOT NULL,
			path TEXT NOT NULL,
			position INTEGER NOT NULL,
			UNIQUE(workspace_id, path),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
	}

	for _, stmt := range statements {
		_, err := db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute migration statement: %w", err)
		}
	}

	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func runMigrations() error {
	var columnExists bool
	err := db.QueryRow(
		"SELECT COUNT(*) > 0 FROM pragma_table_info('commands') WHERE name = 'template_params'",
	).Scan(&columnExists)
	if err != nil {
		return err
	}
	if !columnExists {
		_, err = db.Exec("ALTER TABLE commands ADD COLUMN template_params TEXT DEFAULT ''")
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateWorkspace(workspace *models.Workspace) error {
	ignoredCommandsJSON, err := workspace.GetIgnoredCommandsJSON()
	if err != nil {
		return err
	}

	result, err := db.Exec(
		"INSERT INTO workspaces (name, path, ignored_commands) VALUES (?, ?, ?)",
		workspace.Name, workspace.Path, ignoredCommandsJSON,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	workspace.ID = id
	return nil
}

func GetWorkspaces() ([]models.Workspace, error) {
	rows, err := db.Query("SELECT id, name, path, ignored_commands FROM workspaces ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []models.Workspace
	for rows.Next() {
		var ws models.Workspace
		var ignoredCommandsStr string
		err := rows.Scan(&ws.ID, &ws.Name, &ws.Path, &ignoredCommandsStr)
		if err != nil {
			return nil, err
		}
		ws.SetIgnoredCommandsFromJSON(ignoredCommandsStr)
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func GetWorkspaceByID(id int64) (*models.Workspace, error) {
	var ws models.Workspace
	var ignoredCommandsStr string
	err := db.QueryRow(
		"SELECT id, name, path, ignored_commands FROM workspaces WHERE id = ?",
		id,
	).Scan(&ws.ID, &ws.Name, &ws.Path, &ignoredCommandsStr)
	if err != nil {
		return nil, err
	}
	ws.SetIgnoredCommandsFromJSON(ignoredCommandsStr)
	return &ws, nil
}

func UpdateWorkspace(workspace *models.Workspace) error {
	ignoredCommandsJSON, err := workspace.GetIgnoredCommandsJSON()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE workspaces SET name = ?, path = ?, ignored_commands = ? WHERE id = ?",
		workspace.Name, workspace.Path, ignoredCommandsJSON, workspace.ID,
	)
	return err
}

func DeleteWorkspace(id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM commands WHERE workspace_id = ?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM command_groups WHERE workspace_id = ?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM recent_paths WHERE workspace_id = ?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM workspaces WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetWorkspaceCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM workspaces").Scan(&count)
	return count, err
}

func CreateCommandGroup(group *models.CommandGroup) error {
	result, err := db.Exec(
		"INSERT INTO command_groups (name, workspace_id) VALUES (?, ?)",
		group.Name, group.WorkspaceID,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	group.ID = id
	return nil
}

func GetCommandGroupsByWorkspace(workspaceID int64) ([]models.CommandGroup, error) {
	rows, err := db.Query(
		"SELECT id, name, workspace_id FROM command_groups WHERE workspace_id = ? ORDER BY id DESC",
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.CommandGroup
	for rows.Next() {
		var g models.CommandGroup
		err := rows.Scan(&g.ID, &g.Name, &g.WorkspaceID)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func UpdateCommandGroup(group *models.CommandGroup) error {
	_, err := db.Exec(
		"UPDATE command_groups SET name = ? WHERE id = ?",
		group.Name, group.ID,
	)
	return err
}

func DeleteCommandGroup(id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM commands WHERE group_id = ?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM command_groups WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func CreateCommand(cmd *models.Command) error {
	templateParamsJSON, err := cmd.GetTemplateParamsJSON()
	if err != nil {
		return err
	}

	result, err := db.Exec(
		"INSERT INTO commands (name, content, description, group_id, workspace_id, template_params) VALUES (?, ?, ?, ?, ?, ?)",
		cmd.Name, cmd.Content, cmd.Description, cmd.GroupID, cmd.WorkspaceID, templateParamsJSON,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	cmd.ID = id
	return nil
}

func GetCommandsByWorkspace(workspaceID int64) ([]models.Command, error) {
	rows, err := db.Query(
		"SELECT id, name, content, description, group_id, workspace_id, template_params FROM commands WHERE workspace_id = ? ORDER BY id DESC",
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []models.Command
	for rows.Next() {
		var c models.Command
		var templateParamsStr string
		err := rows.Scan(&c.ID, &c.Name, &c.Content, &c.Description, &c.GroupID, &c.WorkspaceID, &templateParamsStr)
		if err != nil {
			return nil, err
		}
		c.SetTemplateParamsFromJSON(templateParamsStr)
		commands = append(commands, c)
	}
	return commands, nil
}

func CommandExists(workspaceID int64, content string) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM commands WHERE workspace_id = ? AND content = ?",
		workspaceID, content,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateCommand(cmd *models.Command) error {
	templateParamsJSON, err := cmd.GetTemplateParamsJSON()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE commands SET name = ?, content = ?, description = ?, group_id = ?, template_params = ? WHERE id = ?",
		cmd.Name, cmd.Content, cmd.Description, cmd.GroupID, templateParamsJSON, cmd.ID,
	)
	return err
}

func DeleteCommand(id int64) error {
	_, err := db.Exec("DELETE FROM commands WHERE id = ?", id)
	return err
}

func AddRecentPath(workspaceID int64, path string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM recent_paths WHERE workspace_id = ? AND path = ?)",
		workspaceID, path,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		_, err = tx.Exec(
			"UPDATE recent_paths SET position = 0 WHERE workspace_id = ? AND path = ?",
			workspaceID, path,
		)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			"UPDATE recent_paths SET position = position + 1 WHERE workspace_id = ? AND path != ?",
			workspaceID, path,
		)
	} else {
		_, err = tx.Exec(
			"UPDATE recent_paths SET position = position + 1 WHERE workspace_id = ?",
			workspaceID,
		)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			"INSERT INTO recent_paths (workspace_id, path, position) VALUES (?, ?, 0)",
			workspaceID, path,
		)
	}
	if err != nil {
		return err
	}

	var count int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM recent_paths WHERE workspace_id = ?",
		workspaceID,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count > 20 {
		_, err = tx.Exec(
			"DELETE FROM recent_paths WHERE workspace_id = ? AND position >= 20",
			workspaceID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetRecentPaths(workspaceID int64) ([]models.RecentPath, error) {
	rows, err := db.Query(
		"SELECT id, workspace_id, path, position FROM recent_paths WHERE workspace_id = ? ORDER BY position ASC",
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []models.RecentPath
	for rows.Next() {
		var p models.RecentPath
		err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Path, &p.Position)
		if err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func DeleteRecentPath(id int64) error {
	_, err := db.Exec("DELETE FROM recent_paths WHERE id = ?", id)
	return err
}

func ClearRecentPaths(workspaceID int64) error {
	_, err := db.Exec("DELETE FROM recent_paths WHERE workspace_id = ?", workspaceID)
	return err
}

func ExportWorkspace(workspaceID int64) (*models.WorkspaceExport, error) {
	workspace, err := GetWorkspaceByID(workspaceID)
	if err != nil {
		return nil, err
	}

	groups, err := GetCommandGroupsByWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	commands, err := GetCommandsByWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	export := models.NewWorkspaceExport()
	export.Name = workspace.Name
	export.Groups = groups
	export.Commands = commands
	export.Ignored = workspace.IgnoredCommands

	return export, nil
}

func ImportWorkspace(exportData *models.WorkspaceExport, path string) (*models.Workspace, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	workspace := &models.Workspace{
		Name:            exportData.Name,
		Path:            path,
		IgnoredCommands: exportData.Ignored,
	}

	ignoredCommandsJSON, err := workspace.GetIgnoredCommandsJSON()
	if err != nil {
		return nil, err
	}

	result, err := tx.Exec(
		"INSERT INTO workspaces (name, path, ignored_commands) VALUES (?, ?, ?)",
		workspace.Name, workspace.Path, ignoredCommandsJSON,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	workspace.ID = id

	oldToNewGroupID := make(map[int64]int64)
	for _, group := range exportData.Groups {
		groupResult, err := tx.Exec(
			"INSERT INTO command_groups (name, workspace_id) VALUES (?, ?)",
			group.Name, id,
		)
		if err != nil {
			return nil, err
		}
		newGroupID, err := groupResult.LastInsertId()
		if err != nil {
			return nil, err
		}
		oldToNewGroupID[group.ID] = newGroupID
	}

	for _, cmd := range exportData.Commands {
		var newGroupID *int64 = nil
		if cmd.GroupID != nil {
			if gid, ok := oldToNewGroupID[*cmd.GroupID]; ok {
				newGroupID = &gid
			}
		}

		templateParamsJSON, err := cmd.GetTemplateParamsJSON()
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(
			"INSERT INTO commands (name, content, description, group_id, workspace_id, template_params) VALUES (?, ?, ?, ?, ?, ?)",
			cmd.Name, cmd.Content, cmd.Description, newGroupID, id, templateParamsJSON,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return workspace, nil
}

func ExportDatabase() (*models.DatabaseBackup, error) {
	backup := models.NewDatabaseBackup()

	workspaces, err := GetWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("failed to export workspaces: %w", err)
	}
	backup.Workspaces = workspaces

	rows, err := db.Query("SELECT id, name, workspace_id FROM command_groups ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to export command groups: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var g models.CommandGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.WorkspaceID); err != nil {
			return nil, fmt.Errorf("failed to scan command group: %w", err)
		}
		backup.Groups = append(backup.Groups, g)
	}

	cmdRows, err := db.Query("SELECT id, name, content, description, group_id, workspace_id, template_params FROM commands ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to export commands: %w", err)
	}
	defer cmdRows.Close()
	for cmdRows.Next() {
		var c models.Command
		var templateParamsStr string
		if err := cmdRows.Scan(&c.ID, &c.Name, &c.Content, &c.Description, &c.GroupID, &c.WorkspaceID, &templateParamsStr); err != nil {
			return nil, fmt.Errorf("failed to scan command: %w", err)
		}
		c.SetTemplateParamsFromJSON(templateParamsStr)
		backup.Commands = append(backup.Commands, c)
	}

	pathRows, err := db.Query("SELECT id, workspace_id, path, position FROM recent_paths ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to export recent paths: %w", err)
	}
	defer pathRows.Close()
	for pathRows.Next() {
		var p models.RecentPath
		if err := pathRows.Scan(&p.ID, &p.WorkspaceID, &p.Path, &p.Position); err != nil {
			return nil, fmt.Errorf("failed to scan recent path: %w", err)
		}
		backup.RecentPaths = append(backup.RecentPaths, p)
	}

	return backup, nil
}

func DatabaseHasData() (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM workspaces").Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func ImportDatabase(backup *models.DatabaseBackup) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM recent_paths"); err != nil {
		return fmt.Errorf("failed to clear recent_paths: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM commands"); err != nil {
		return fmt.Errorf("failed to clear commands: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM command_groups"); err != nil {
		return fmt.Errorf("failed to clear command_groups: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM workspaces"); err != nil {
		return fmt.Errorf("failed to clear workspaces: %w", err)
	}

	oldToNewWorkspaceID := make(map[int64]int64)
	for _, ws := range backup.Workspaces {
		ignoredCommandsJSON, err := ws.GetIgnoredCommandsJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal ignored commands for workspace %s: %w", ws.Name, err)
		}
		result, err := tx.Exec(
			"INSERT INTO workspaces (name, path, ignored_commands) VALUES (?, ?, ?)",
			ws.Name, ws.Path, ignoredCommandsJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert workspace %s: %w", ws.Name, err)
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get new workspace id: %w", err)
		}
		oldToNewWorkspaceID[ws.ID] = newID
	}

	oldToNewGroupID := make(map[int64]int64)
	for _, g := range backup.Groups {
		newWorkspaceID, ok := oldToNewWorkspaceID[g.WorkspaceID]
		if !ok {
			continue
		}
		result, err := tx.Exec(
			"INSERT INTO command_groups (name, workspace_id) VALUES (?, ?)",
			g.Name, newWorkspaceID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert command group %s: %w", g.Name, err)
		}
		newGroupID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get new group id: %w", err)
		}
		oldToNewGroupID[g.ID] = newGroupID
	}

	for _, cmd := range backup.Commands {
		newWorkspaceID, ok := oldToNewWorkspaceID[cmd.WorkspaceID]
		if !ok {
			continue
		}
		var newGroupID *int64
		if cmd.GroupID != nil {
			if gid, exists := oldToNewGroupID[*cmd.GroupID]; exists {
				newGroupID = &gid
			}
		}
		templateParamsJSON, err := cmd.GetTemplateParamsJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal template params for command %s: %w", cmd.Name, err)
		}
		_, err = tx.Exec(
			"INSERT INTO commands (name, content, description, group_id, workspace_id, template_params) VALUES (?, ?, ?, ?, ?, ?)",
			cmd.Name, cmd.Content, cmd.Description, newGroupID, newWorkspaceID, templateParamsJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert command %s: %w", cmd.Name, err)
		}
	}

	for _, rp := range backup.RecentPaths {
		newWorkspaceID, ok := oldToNewWorkspaceID[rp.WorkspaceID]
		if !ok {
			continue
		}
		_, err := tx.Exec(
			"INSERT INTO recent_paths (workspace_id, path, position) VALUES (?, ?, ?)",
			newWorkspaceID, rp.Path, rp.Position,
		)
		if err != nil {
			return fmt.Errorf("failed to insert recent path %s: %w", rp.Path, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
