package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
)

var nonIntRegex, _ = regexp.Compile("[^\\d]+")

func hasMigrationFileBeenApplied(db *sql.DB, fileId int) (ran bool, err error) {
	statement, err := db.Prepare(fmt.Sprintf("SELECT COUNT(*) > 0 FROM %s WHERE file_id = ? AND applied = ?", SimpleDbMigrationTableName))
	if err != nil {
		return
	}
	row := statement.QueryRow(fileId, true)
	err = row.Scan(&ran)
	return
}

func createSimpleDbMigrationTable(db *sql.DB) (err error) {
	statement := fmt.Sprintf("CREATE TABLE %s (id INTEGER NOT NULL PRIMARY KEY, file_id INTEGER NOT NULL UNIQUE, applied BOOLEAN NOT NULL)", SimpleDbMigrationTableName)
	_, err = db.Exec(statement)
	if err != nil {
		return err
	}
	return
}

func getConcreteMigratableDb(dbType string) (m MigratableDb, err error) {
	if dbType == "sqlite" {
		m = &MigratableSqlite{}
	} else {
		err = errors.New(fmt.Sprintf("%s is not supported", dbType))
	}
	return
}

func migrateFile(migratableDb MigratableDb, db *sql.DB, sqlFilePath string, up bool) {
	sqlFileName := filepath.Base(sqlFilePath)
	sqlFileId, err := strconv.Atoi(nonIntRegex.ReplaceAllString(sqlFileName, ""))
	if err != nil {
		log.Fatal(err)
	}

	if up {
		isFileApplied, err := hasMigrationFileBeenApplied(db, sqlFileId)
		if err != nil {
			log.Fatal(err)
		}

		if isFileApplied {
			log.Printf("%s has been ran, continuing", sqlFileName)
			return
		}
	}

	sqlFileContent, err := readFileIntoString(sqlFilePath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(sqlFileContent)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("executed %s: %s", sqlFilePath, sqlFileContent)

	err = migratableDb.recordMigration(db, sqlFileId, up)
	if err != nil {
		log.Fatal(err)
	}
}

func upAll(migratableDb MigratableDb, db *sql.DB, deltasDir string) {
	sqlFilePaths, err := filepath.Glob(filepath.Join(deltasDir, "*.up.sql"))
	if err != nil {
		log.Fatal(err)
	}

	for _, sqlFilePath := range sqlFilePaths {
		migrateFile(migratableDb, db, sqlFilePath, true)
	}
}

func downOne(migratableDb MigratableDb, db *sql.DB, deltasDir string, id int) {
	pattern := fmt.Sprintf("%d.*.down.sql", id)
	sqlFilePaths, err := filepath.Glob(filepath.Join(deltasDir, pattern))
	if err != nil {
		log.Fatal(err)
	}

	if len(sqlFilePaths) == 0 {
		log.Fatal(errors.New(fmt.Sprintf("no file matching the pattern: %s", pattern)))
	} else if len(sqlFilePaths) > 1 {
		log.Fatal(errors.New(fmt.Sprintf("more than one file matching the pattern: %s", pattern)))
	}

	migrateFile(migratableDb, db, sqlFilePaths[0], false)
}

func main() {
	jsonConfigPathRef := flag.String("config", "./simple-db-migrate.json", "Path to the configuration json file.")
	downIdRef := flag.Int("down", -1, "Run the corresponding down migration SQL")
	flag.Parse()

	jsonConfigPath := *jsonConfigPathRef
	downId := *downIdRef

	configJson, err := parseConfigJson(jsonConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	migratableDb, err := getConcreteMigratableDb(configJson.Type)
	if err != nil {
		log.Fatal(err)
	}

	rootDir := filepath.Dir(jsonConfigPath)
	db, err := migratableDb.getDbFromConfig(rootDir, configJson)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createSimpleDbMigrationTable(db)
	if err != nil {
		log.Printf("%v, continuing", err)
	}

	deltasDir := filepath.Join(rootDir, configJson.Deltas)

	if downId > 0 {
		downOne(migratableDb, db, deltasDir, downId)
	} else {
		upAll(migratableDb, db, deltasDir)
	}

	log.Println("finish running migrations")
}
