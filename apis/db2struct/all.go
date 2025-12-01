package db2struct

import (
	"database/sql"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/stoewer/go-strcase"

	_ "github.com/go-sql-driver/mysql"
)

type Table struct {
	Name    string
	Comment string
}

func getTables(db *sql.DB, dbName string, tableName string) (ts []Table, err error) {
	sqlCommand := `SELECT TABLE_NAME, TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ?`
	args := []interface{}{dbName}
	if len(tableName) > 0 {
		sqlCommand += " AND TABLE_NAME = ?"
		args = append(args, tableName)
	}
	r, err := db.Query(sqlCommand, args...)
	if err != nil {
		return
	}
	// nolint
	defer r.Close()
	for r.Next() {
		var table Table
		if err = r.Scan(&table.Name, &table.Comment); err != nil {
			return
		}
		ts = append(ts, table)
	}
	return
}

func GenAll(dir string, dbConfig DBConfig, options ...Option) (err error) {
	db, err := GetConnection(dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB)
	if err != nil {
		return
	}

	tables, err := getTables(db, dbConfig.DB, "")
	if err != nil {
		return
	}

	for _, table := range tables {
		var ret []byte
		ret, err = Gen(table.Name, dbConfig, options...)
		if err != nil {
			return err
		}

		if err = helpers.ImportAndWrite(ret, filepath.Join(dir, strcase.SnakeCase(table.Name)+".go")); err != nil {
			return err
		}
	}
	return
}
