package db

// 提供MySQL数据库适配器实现.

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spelens-gud/gsus/internal/errors"
)

// MySQLAdapter struct    MySQL适配器.
type MySQLAdapter struct {
	db *sql.DB
}

// Connect method    连接MySQL数据库.
func (a *MySQLAdapter) Connect(ctx context.Context, config *Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		getCharset(config.Charset, "utf8mb4"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("连接MySQL失败: %s", err))
	}

	if err = db.PingContext(ctx); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("MySQL连接测试失败: %s", err))
	}

	a.db = db
	return nil
}

// Close method    关闭MySQL连接.
func (a *MySQLAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// GetTables method    获取MySQL所有表.
func (a *MySQLAdapter) GetTables(ctx context.Context, database string, tableFilter string) ([]Table, error) {
	query := `SELECT TABLE_NAME, TABLE_COMMENT
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = ?`
	args := []interface{}{database}

	if tableFilter != "" {
		query += " AND TABLE_NAME = ?"
		args = append(args, tableFilter)
	}

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询表列表失败: %s", err))
	}
	//nolint:errcheck
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.Name, &table.Comment); err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
				fmt.Sprintf("扫描表信息失败: %s", err))
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// GetColumns method    获取MySQL表的所有列.
func (a *MySQLAdapter) GetColumns(ctx context.Context, database, table string) ([]Column, error) {
	query := `SELECT
		COLUMN_NAME,
		DATA_TYPE,
		IS_NULLABLE,
		COLUMN_KEY,
		EXTRA,
		COLUMN_DEFAULT,
		COLUMN_COMMENT,
		COLUMN_TYPE
	FROM INFORMATION_SCHEMA.COLUMNS
	WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	ORDER BY ORDINAL_POSITION`

	rows, err := a.db.QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询列信息失败: %s", err))
	}
	//nolint:errcheck
	defer rows.Close()

	var columns []Column
	typeMap := a.TypeMapping()

	for rows.Next() {
		var (
			columnName    string
			dataType      string
			isNullable    string
			columnKey     string
			extra         string
			columnDefault sql.NullString
			columnComment string
			columnType    string
		)

		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnKey,
			&extra, &columnDefault, &columnComment, &columnType); err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
				fmt.Sprintf("扫描列信息失败: %s", err))
		}

		column := Column{
			Name:         columnName,
			Type:         columnType,
			GoType:       mapMySQLTypeToGo(dataType, isNullable == "YES", typeMap),
			Nullable:     isNullable == "YES",
			IsPrimaryKey: columnKey == "PRI",
			IsAutoIncr:   extra == "auto_increment",
			Comment:      columnComment,
			Extra:        extra,
		}

		if columnDefault.Valid {
			column.Default = columnDefault.String
		}

		columns = append(columns, column)
	}

	return columns, nil
}

// GetIndexes method    获取MySQL表的所有索引.
func (a *MySQLAdapter) GetIndexes(ctx context.Context, database, table string) ([]Index, error) {
	query := `SELECT
		INDEX_NAME,
		COLUMN_NAME,
		NON_UNIQUE,
		INDEX_TYPE
	FROM INFORMATION_SCHEMA.STATISTICS
	WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	ORDER BY INDEX_NAME, SEQ_IN_INDEX`

	rows, err := a.db.QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询索引信息失败: %s", err))
	}
	//nolint:errcheck
	defer rows.Close()

	indexMap := make(map[string]*Index)

	for rows.Next() {
		var (
			keyName    string
			columnName string
			nonUnique  int
			indexType  string
		)

		if err := rows.Scan(&keyName, &columnName, &nonUnique, &indexType); err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
				fmt.Sprintf("扫描索引信息失败: %s", err))
		}

		if idx, exists := indexMap[keyName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[keyName] = &Index{
				Name:       keyName,
				Columns:    []string{columnName},
				IsUnique:   nonUnique == 0,
				IsPrimary:  keyName == "PRIMARY",
				IsFullText: indexType == "FULLTEXT",
			}
		}
	}

	var indexes []Index
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}

	return indexes, nil
}

// TypeMapping method    获取MySQL类型映射.
func (a *MySQLAdapter) TypeMapping() map[string]string {
	return map[string]string{
		"tinyint":    "int",
		"int":        "int",
		"smallint":   "int",
		"mediumint":  "int",
		"bigint":     "int64",
		"char":       "string",
		"enum":       "string",
		"varchar":    "string",
		"longtext":   "string",
		"mediumtext": "string",
		"text":       "string",
		"tinytext":   "string",
		"date":       "time.Time",
		"datetime":   "time.Time",
		"time":       "time.Time",
		"timestamp":  "time.Time",
		"decimal":    "float64",
		"double":     "float64",
		"float":      "float32",
		"binary":     "[]byte",
		"blob":       "[]byte",
		"longblob":   "[]byte",
		"mediumblob": "[]byte",
		"varbinary":  "[]byte",
		"json":       "interface{}",
	}
}

// mapMySQLTypeToGo function    映射MySQL类型到Go类型.
func mapMySQLTypeToGo(mysqlType string, nullable bool, typeMap map[string]string) string {
	if nullable {
		switch mysqlType {
		case "tinyint", "int", "smallint", "mediumint", "bigint":
			return "sql.NullInt64"
		case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
			return "sql.NullString"
		case "date", "datetime", "time", "timestamp":
			return "sql.NullTime"
		case "decimal", "double", "float":
			return "sql.NullFloat64"
		}
	}

	if goType, exists := typeMap[mysqlType]; exists {
		return goType
	}

	return "interface{}"
}

// getCharset function    获取字符集，如果为空则使用默认值.
func getCharset(charset, defaultCharset string) string {
	if charset == "" {
		return defaultCharset
	}
	return charset
}
