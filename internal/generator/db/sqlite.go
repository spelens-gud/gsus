package db

// 提供SQLite数据库适配器实现.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spelens-gud/gsus/internal/errors"
)

// SQLiteAdapter struct    SQLite适配器.
type SQLiteAdapter struct {
	db *sql.DB
}

// Connect method    连接SQLite数据库.
func (a *SQLiteAdapter) Connect(ctx context.Context, config *Config) error {
	// SQLite使用文件路径作为数据库名
	dsn := config.Database
	if dsn == "" {
		dsn = config.Host // 兼容使用Host字段存储文件路径
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("连接SQLite失败: %s", err))
	}

	if err = db.PingContext(ctx); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("SQLite连接测试失败: %s", err))
	}

	a.db = db
	return nil
}

// Close method    关闭SQLite连接.
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// GetTables method    获取SQLite所有表.
func (a *SQLiteAdapter) GetTables(ctx context.Context, database string, tableFilter string) ([]Table, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table'`
	args := []interface{}{}

	if tableFilter != "" {
		query += " AND name = ?"
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
		if err := rows.Scan(&table.Name); err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
				fmt.Sprintf("扫描表信息失败: %s", err))
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// GetColumns method    获取SQLite表的所有列.
func (a *SQLiteAdapter) GetColumns(ctx context.Context, database, table string) ([]Column, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", table)

	rows, err := a.db.QueryContext(ctx, query)
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
			cid          int
			name         string
			dataType     string
			notNull      int
			defaultValue sql.NullString
			pk           int
		)

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
				fmt.Sprintf("扫描列信息失败: %s", err))
		}

		column := Column{
			Name:         name,
			Type:         dataType,
			GoType:       mapSQLiteTypeToGo(dataType, notNull == 0, typeMap),
			Nullable:     notNull == 0,
			IsPrimaryKey: pk == 1,
		}

		if defaultValue.Valid {
			column.Default = defaultValue.String
		}

		// SQLite的自增需要额外检查
		if pk == 1 {
			column.IsAutoIncr = a.isAutoIncrement(ctx, table, name)
		}

		columns = append(columns, column)
	}

	return columns, nil
}

// isAutoIncrement method    检查列是否为自增.
func (a *SQLiteAdapter) isAutoIncrement(ctx context.Context, table, column string) bool {
	query := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", table)
	var createSQL string
	if err := a.db.QueryRowContext(ctx, query).Scan(&createSQL); err != nil {
		return false
	}
	return strings.Contains(strings.ToUpper(createSQL), "AUTOINCREMENT")
}

// GetIndexes method    获取SQLite表的所有索引.
func (a *SQLiteAdapter) GetIndexes(ctx context.Context, database, table string) ([]Index, error) {
	query := fmt.Sprintf("PRAGMA index_list(%s)", table)

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询索引信息失败: %s", err))
	}
	//nolint:errcheck
	defer rows.Close()

	var indexes []Index

	for rows.Next() {
		var (
			seq     int
			name    string
			unique  int
			origin  string
			partial int
		)

		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
				fmt.Sprintf("扫描索引信息失败: %s", err))
		}

		// 获取索引的列信息
		columns, err := a.getIndexColumns(ctx, name)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, Index{
			Name:      name,
			Columns:   columns,
			IsUnique:  unique == 1,
			IsPrimary: origin == "pk",
		})
	}

	return indexes, nil
}

// getIndexColumns method    获取索引的列.
func (a *SQLiteAdapter) getIndexColumns(ctx context.Context, indexName string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA index_info(%s)", indexName)

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询索引列信息失败: %s", err))
	}
	//nolint:errcheck
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var (
			seqno int
			cid   int
			name  string
		)

		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
				fmt.Sprintf("扫描索引列信息失败: %s", err))
		}

		columns = append(columns, name)
	}

	return columns, nil
}

// TypeMapping method    获取SQLite类型映射.
func (a *SQLiteAdapter) TypeMapping() map[string]string {
	return map[string]string{
		"INTEGER":   "int64",
		"INT":       "int",
		"TINYINT":   "int",
		"SMALLINT":  "int",
		"MEDIUMINT": "int",
		"BIGINT":    "int64",
		"TEXT":      "string",
		"VARCHAR":   "string",
		"CHAR":      "string",
		"REAL":      "float64",
		"DOUBLE":    "float64",
		"FLOAT":     "float32",
		"NUMERIC":   "float64",
		"DECIMAL":   "float64",
		"BOOLEAN":   "bool",
		"BOOL":      "bool",
		"DATE":      "time.Time",
		"DATETIME":  "time.Time",
		"TIMESTAMP": "time.Time",
		"BLOB":      "[]byte",
	}
}

// sqliteNullableTypeMap 定义SQLite可空类型到sql.Null*类型的映射.
var sqliteNullableTypeMap = map[string]func(string) bool{
	"sql.NullInt64": func(sqliteType string) bool {
		return strings.Contains(sqliteType, "INT")
	},
	"sql.NullString": func(sqliteType string) bool {
		return strings.Contains(sqliteType, "TEXT") || strings.Contains(sqliteType, "CHAR")
	},
	"sql.NullBool": func(sqliteType string) bool {
		return strings.Contains(sqliteType, "BOOL")
	},
	"sql.NullFloat64": func(sqliteType string) bool {
		return strings.Contains(sqliteType, "REAL") ||
			strings.Contains(sqliteType, "FLOAT") ||
			strings.Contains(sqliteType, "DOUBLE") ||
			strings.Contains(sqliteType, "NUMERIC") ||
			strings.Contains(sqliteType, "DECIMAL")
	},
	"sql.NullTime": func(sqliteType string) bool {
		return strings.Contains(sqliteType, "DATE") || strings.Contains(sqliteType, "TIME")
	},
}

// mapSQLiteTypeToGo function    映射SQLite类型到Go类型.
func mapSQLiteTypeToGo(sqliteType string, nullable bool, typeMap map[string]string) string {
	sqliteType = strings.ToUpper(sqliteType)

	if nullable {
		for nullType, condition := range sqliteNullableTypeMap {
			if condition(sqliteType) {
				return nullType
			}
		}
	}

	// 精确匹配
	if goType, exists := typeMap[sqliteType]; exists {
		return goType
	}

	// 模糊匹配
	for dbType, goType := range typeMap {
		if strings.Contains(sqliteType, dbType) {
			return goType
		}
	}

	return "interface{}"
}
