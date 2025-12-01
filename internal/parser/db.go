package parser

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/spelens-gud/gsus/internal/errors"
)

type CTypes struct {
	Key  string
	Info map[string]string
}

func GetConnection(mariadbUser string, mariadbPassword string, mariadbHost string,
	mariadbPort int, mariadbDatabase string) (db *sql.DB, err error) {
	if mariadbPassword != "" {
		db, err = sql.Open("mysql", mariadbUser+":"+
			mariadbPassword+"@tcp("+mariadbHost+":"+strconv.Itoa(mariadbPort)+")/"+mariadbDatabase+"?&parseTime=True")
	} else {
		db, err = sql.Open("mysql", mariadbUser+"@tcp("+mariadbHost+":"+
			strconv.Itoa(mariadbPort)+")/"+mariadbDatabase+"?&parseTime=True")
	}

	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase, fmt.Sprintf("无法打开数据库连接: %s", err))
	}
	return
}

func GetColumnsFromMysqlTable(db *sql.DB, dbName, table string) (types *[]CTypes, err error) {
	// Store column as map of maps
	var columnDataTypes []CTypes
	// Select columnd data from INFORMATION_SCHEMA
	columnDataTypeQuery := "SELECT COLUMN_NAME, COLUMN_KEY, DATA_TYPE, IS_NULLABLE,COLUMN_COMMENT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ?"

	rows, err := db.Query(columnDataTypeQuery, dbName, table)

	if err != nil {
		fmt.Println("Error selecting from db: " + err.Error())
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase, fmt.Sprintf("无法从数据库中获取列数据: %s", err))
	}
	if rows != nil {
		// nolint
		defer rows.Close()
	} else {
		return nil, errors.New(errors.ErrCodeDatabase, "表未返回任何结果")
	}

	for rows.Next() {
		var column string
		var columnKey string
		var dataType string
		var nullable string
		var comment string
		_ = rows.Scan(&column, &columnKey, &dataType, &nullable, &comment)
		columnDataTypes = append(columnDataTypes, CTypes{
			Key:  column,
			Info: map[string]string{"value": dataType, "nullable": nullable, "primary": columnKey, "comment": comment},
		})
	}

	return &columnDataTypes, err
}
