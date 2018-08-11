package main

import "database/sql"

type MigratableDb interface {
	getDbFromConfig(configPath string, configJson *ConfigJson) (*sql.DB, error)
	recordMigration(db *sql.DB, fileId int, applied bool) (err error)
}
