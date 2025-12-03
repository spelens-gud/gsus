package db

// 提供MySQL数据库适配器实现.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

	// 按索引名称排序，确保输出顺序稳定
	var indexNames []string
	for name := range indexMap {
		indexNames = append(indexNames, name)
	}

	// 对索引名称排序
	for i := 0; i < len(indexNames); i++ {
		for j := i + 1; j < len(indexNames); j++ {
			if indexNames[i] > indexNames[j] {
				indexNames[i], indexNames[j] = indexNames[j], indexNames[i]
			}
		}
	}

	var indexes []Index
	for _, name := range indexNames {
		indexes = append(indexes, *indexMap[name])
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

// BuildGormTag method    构建MySQL的GORM标签.
func (a *MySQLAdapter) BuildGormTag(col Column, indexes []Index) string {
	var parts []string

	// 列名
	parts = append(parts, "column:"+col.Name)

	// 主键
	if col.IsPrimaryKey {
		parts = append(parts, "primaryKey")
	}

	// 自增
	if col.IsAutoIncr {
		parts = append(parts, "autoIncrement")
	}

	// 检查是否为时间字段，添加自动时间戳标签
	if isTimeColumn(col) {
		if isCreateTimeColumn(col.Name) {
			parts = append(parts, "autoCreateTime")
		} else if isUpdateTimeColumn(col.Name) {
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
	if col.Default != "" && !isCreateTimeColumn(col.Name) && !isUpdateTimeColumn(col.Name) {
		parts = append(parts, "default:"+col.Default)
	}

	// 注释
	if col.Comment != "" {
		parts = append(parts, "comment:"+col.Comment)
	}

	// 处理索引
	parts = append(parts, buildMySQLIndexTags(col.Name, indexes)...)

	return strings.Join(parts, ";")
}

// buildMySQLIndexTags function    构建MySQL索引标签.
func buildMySQLIndexTags(columnName string, indexes []Index) []string {
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

// isCreateTimeColumn function    判断是否为创建时间字段.
func isCreateTimeColumn(columnName string) bool {
	lowerName := strings.ToLower(columnName)
	return lowerName == "created_at" ||
		lowerName == "create_time" ||
		lowerName == "createtime" ||
		lowerName == "created_time" ||
		lowerName == "createdtime" ||
		lowerName == "gmt_create" ||
		lowerName == "gmtcreate"
}

// isUpdateTimeColumn function    判断是否为更新时间字段.
func isUpdateTimeColumn(columnName string) bool {
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

// isTimeColumn function    判断是否为时间类型字段.
func isTimeColumn(col Column) bool {
	// 检查Go类型
	if strings.Contains(col.GoType, "time.Time") ||
		strings.Contains(col.GoType, "sql.NullTime") {
		return true
	}

	// 检查数据库类型
	lowerType := strings.ToLower(col.Type)
	return strings.Contains(lowerType, "datetime") ||
		strings.Contains(lowerType, "timestamp") ||
		lowerType == "date" ||
		lowerType == "time"
}
