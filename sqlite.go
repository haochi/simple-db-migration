package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
)

type MigratableSqlite struct {
}

func (m *MigratableSqlite) getDbFromConfig(rootDir string, configJson *ConfigJson) (db *sql.DB, err error) {
	sqliteDbPath := filepath.Join(rootDir, configJson.Sqlite.File)
	db, err = sql.Open("sqlite3", sqliteDbPath)
	return
}

func (m *MigratableSqlite) recordMigration(db *sql.DB, fileId int, applied bool) (err error) {
	statement, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (file_id, applied) VALUES (?, ?) ON CONFLICT(file_id) DO UPDATE SET applied = excluded.applied", SimpleDbMigrationTableName))
	if err != nil {
		return
	}
	_, err = statement.Exec(fileId, applied)
	return
}
