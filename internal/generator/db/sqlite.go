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
	query := fmt.Sprintf("SELECT mysql FROM sqlite_master WHERE type='table' AND name='%s'", table)
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

// BuildGormTag method    构建SQLite的GORM标签.
func (a *SQLiteAdapter) BuildGormTag(col Column, indexes []Index) string {
	var parts []string

	// 列名
	parts = append(parts, "column:"+col.Name)

	// 主键
	if col.IsPrimaryKey {
		parts = append(parts, "primaryKey")
	}

	// 自增（SQLite使用INTEGER PRIMARY KEY会自动自增）
	if col.IsAutoIncr {
		parts = append(parts, "autoIncrement")
	}

	// 检查是否为时间字段，添加自动时间戳标签
	if isSQLiteTimeColumn(col) {
		if isSQLiteCreateTimeColumn(col.Name) {
			parts = append(parts, "autoCreateTime")
		} else if isSQLiteUpdateTimeColumn(col.Name) {
			parts = append(parts, "autoUpdateTime")
		}
	}

	// 类型
	parts = append(parts, "type:"+col.Type)

	// 非空
	if !col.Nullable {
		parts = append(parts, "not null")
	}

	// 默认值（如果已经有autoCreateTime或autoUpdateTime，可能不需要default）
	if col.Default != "" && !isSQLiteCreateTimeColumn(col.Name) && !isSQLiteUpdateTimeColumn(col.Name) {
		parts = append(parts, "default:"+col.Default)
	}

	// 注释（SQLite不原生支持注释，但GORM可以使用）
	if col.Comment != "" {
		parts = append(parts, "comment:"+col.Comment)
	}

	// 处理索引
	parts = append(parts, buildSQLiteIndexTags(col.Name, indexes)...)

	return strings.Join(parts, ";")
}

// buildSQLiteIndexTags function    构建SQLite索引标签.
func buildSQLiteIndexTags(columnName string, indexes []Index) []string {
	var (
		normalIndexes []string
		uniqueIndexes []string
	)

	for _, idx := range indexes {
		if idx.IsPrimary {
			continue
		}

		// 检查索引是否包含当前列
		hasColumn := false
		for _, col := range idx.Columns {
			if col == columnName {
				hasColumn = true
				break
			}
		}

		if !hasColumn {
			continue
		}

		if idx.IsUnique {
			uniqueIndexes = append(uniqueIndexes, idx.Name)
		} else {
			normalIndexes = append(normalIndexes, idx.Name)
		}
	}

	var parts []string
	if len(normalIndexes) > 0 {
		parts = append(parts, "index:"+strings.Join(normalIndexes, ","))
	}
	if len(uniqueIndexes) > 0 {
		parts = append(parts, "uniqueIndex:"+strings.Join(uniqueIndexes, ","))
	}

	return parts
}

// isSQLiteCreateTimeColumn function    判断是否为创建时间字段.
func isSQLiteCreateTimeColumn(columnName string) bool {
	lowerName := strings.ToLower(columnName)
	return lowerName == "created_at" ||
		lowerName == "create_time" ||
		lowerName == "createtime" ||
		lowerName == "created_time" ||
		lowerName == "createdtime" ||
		lowerName == "gmt_create" ||
		lowerName == "gmtcreate"
}

// isSQLiteUpdateTimeColumn function    判断是否为更新时间字段.
func isSQLiteUpdateTimeColumn(columnName string) bool {
	lowerName := strings.ToLower(columnName)
	return lowerName == "updated_at" ||
		lowerName == "update_time" ||
		lowerName == "updatetime" ||
		lowerName == "updated_time" ||
		lowerName == "updatedtime" ||
		lowerName == "gmt_modified" ||
		lowerName == "gmtmodified" ||
		lowerName == "modified_at" ||
		lowerName == "modify_time"
}

// isSQLiteTimeColumn function    判断是否为时间类型字段.
func isSQLiteTimeColumn(col Column) bool {
	// 检查Go类型
	if strings.Contains(col.GoType, "time.Time") ||
		strings.Contains(col.GoType, "sql.NullTime") {
		return true
	}

	// 检查数据库类型
	upperType := strings.ToUpper(col.Type)
	return strings.Contains(upperType, "DATETIME") ||
		strings.Contains(upperType, "TIMESTAMP") ||
		strings.Contains(upperType, "DATE") ||
		strings.Contains(upperType, "TIME")
}
