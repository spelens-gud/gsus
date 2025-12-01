package db2struct

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// GetColumnsFromMysqlTable Select column details from information schema and return map of map
func GetConnection(mariadbUser string, mariadbPassword string, mariadbHost string, mariadbPort int, mariadbDatabase string) (db *sql.DB, err error) {
	if mariadbPassword != "" {
		db, err = sql.Open("mysql", mariadbUser+":"+mariadbPassword+"@tcp("+mariadbHost+":"+strconv.Itoa(mariadbPort)+")/"+mariadbDatabase+"?&parseTime=True")
	} else {
		db, err = sql.Open("mysql", mariadbUser+"@tcp("+mariadbHost+":"+strconv.Itoa(mariadbPort)+")/"+mariadbDatabase+"?&parseTime=True")
	}
	return
}

func GetColumnsFromMysqlTable(db *sql.DB, dbName, table string) (types *[]cTypes, err error) {
	// Store column as map of maps
	var columnDataTypes []cTypes
	// Select columnd data from INFORMATION_SCHEMA
	columnDataTypeQuery := "SELECT COLUMN_NAME, COLUMN_KEY, DATA_TYPE, IS_NULLABLE,COLUMN_COMMENT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ?"

	rows, err := db.Query(columnDataTypeQuery, dbName, table)

	if err != nil {
		fmt.Println("Error selecting from db: " + err.Error())
		return nil, err
	}
	if rows != nil {
		// nolint
		defer rows.Close()
	} else {
		return nil, errors.New("no results returned for table")
	}

	for rows.Next() {
		var column string
		var columnKey string
		var dataType string
		var nullable string
		var comment string
		_ = rows.Scan(&column, &columnKey, &dataType, &nullable, &comment)
		columnDataTypes = append(columnDataTypes, cTypes{
			Key:  column,
			Info: map[string]string{"value": dataType, "nullable": nullable, "primary": columnKey, "comment": comment},
		})
	}

	return &columnDataTypes, err
}

// Generate go struct entries for a map[string]interface{} structure
func generateMysqlTypes(objs []cTypes, depth int, opts *opt) (string, map[string]string) {
	structure := "struct {"
	fieldNameMap := make(map[string]string)
	for _, obj := range objs {
		mysqlType := obj.Info
		nullable := false
		if mysqlType["nullable"] == "YES" {
			nullable = true
		}

		primary := ""
		if mysqlType["primary"] == "PRI" {
			primary = ";primary_key"
		}

		// Get the corresponding go value type for this mysql type
		var valueType string
		// If the guregu (https://github.com/guregu/null) CLI option is passed use its types, otherwise use go's sql.NullX

		valueType = mysqlTypeToGoType(mysqlType["value"], nullable)
		if rp := opts.typeReplace[valueType]; len(rp) > 0 {
			valueType = rp
		}
		fieldName := fmtFieldName(stringifyFirstChar(obj.Key))
		fieldNameMap[fieldName] = obj.Key
		var annotations []string
		if opts.gormAnnotation {
			annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s%s\"", obj.Key, primary))
		}
		if len(annotations) > 0 {
			structure += fmt.Sprintf("\n%s %s `%s`",
				fieldName,
				valueType,
				strings.Join(annotations, " "))

		} else {
			structure += fmt.Sprintf("\n%s %s",
				fieldName,
				valueType)
		}
		if (opts.commentOutside || !opts.sqlInfo) && len(obj.Info["comment"]) > 0 {
			structure += " // " + obj.Info["comment"]
		}
	}
	structure += "\n}"
	return structure, fieldNameMap
}

// mysqlTypeToGoType converts the mysql types to go compatible sql.Nullable (https://golang.org/pkg/database/sql/) types
func mysqlTypeToGoType(mysqlType string, nullable bool) string {
	switch mysqlType {
	case "tinyint", "int", "smallint", "mediumint":
		if nullable {
			return sqlNullInt
		}
		return golangInt
	case "json":
		return golangInterface
	case "bigint":
		if nullable {
			return sqlNullInt
		}
		return golangInt64
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
		if nullable {
			return sqlNullString
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		if nullable {
			return sqlNullTime
		}
		return golangTime
	case "decimal", "double":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat64
	case "float":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat32
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray
	}
	return golangInterface
}
