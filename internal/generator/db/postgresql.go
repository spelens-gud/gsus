package db

// 提供PostgreSQL数据库适配器实现.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/spelens-gud/gsus/internal/errors"
)

// PostgreSQLAdapter struct    PostgreSQL适配器.
type PostgreSQLAdapter struct {
	db *sql.DB
}

// Connect method    连接PostgreSQL数据库.
func (a *PostgreSQLAdapter) Connect(ctx context.Context, config *Config) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("连接PostgreSQL失败: %s", err))
	}

	if err = db.PingContext(ctx); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("PostgreSQL连接测试失败: %s", err))
	}

	a.db = db
	return nil
}

// Close method    关闭PostgreSQL连接.
func (a *PostgreSQLAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// GetTables method    获取PostgreSQL所有表.
func (a *PostgreSQLAdapter) GetTables(ctx context.Context, database string, tableFilter string) ([]Table, error) {
	query := `SELECT
		table_name,
		COALESCE(obj_description((quote_ident(table_schema)||'.'||quote_ident(table_name))::regclass), '') as comment
	FROM information_schema.tables
	WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`

	args := []interface{}{}
	if tableFilter != "" {
		query += " AND table_name = $1"
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
		table.Schema = "public"
		tables = append(tables, table)
	}

	return tables, nil
}

// GetColumns method    获取PostgreSQL表的所有列.
func (a *PostgreSQLAdapter) GetColumns(ctx context.Context, database, table string) ([]Column, error) {
	query := a.getColumnQuery()
	rows, err := a.db.QueryContext(ctx, query, table)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询列信息失败: %s", err))
	}
	//nolint:errcheck
	defer rows.Close()

	var columns []Column
	typeMap := a.TypeMapping()

	for rows.Next() {
		column, err := a.scanColumn(rows, typeMap)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns, nil
}

// scanColumn method    扫描单行列信息.
func (a *PostgreSQLAdapter) scanColumn(rows *sql.Rows, typeMap map[string]string) (Column, error) {
	var (
		columnName    string
		dataType      string
		isNullable    string
		columnDefault sql.NullString
		columnComment string
		udtName       string
		isPrimary     bool
	)

	if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault,
		&columnComment, &udtName, &isPrimary); err != nil {
		return Column{}, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("扫描列信息失败: %s", err))
	}

	column := Column{
		Name:         columnName,
		Type:         dataType,
		GoType:       mapPostgreSQLTypeToGo(udtName, isNullable == "YES", typeMap),
		Nullable:     isNullable == "YES",
		IsPrimaryKey: isPrimary,
		Comment:      columnComment,
	}

	if columnDefault.Valid {
		column.Default = columnDefault.String
		// 检查是否为序列（自增）
		if len(column.Default) > 7 && column.Default[:7] == "nextval" {
			column.IsAutoIncr = true
		}
	}

	return column, nil
}

// getColumnQuery method    返回获取列信息的SQL查询语句.
func (a *PostgreSQLAdapter) getColumnQuery() string {
	return `SELECT
		c.column_name,
		c.data_type,
		c.is_nullable,
		c.column_default,
		COALESCE(pgd.description, '') as column_comment,
		c.udt_name,
		CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary
	FROM information_schema.columns c
	LEFT JOIN pg_catalog.pg_description pgd
		ON pgd.objoid = (SELECT oid FROM pg_class WHERE relname = c.table_name AND relnamespace =
			(SELECT oid FROM pg_namespace WHERE nspname = c.table_schema))
		AND pgd.objsubid = c.ordinal_position
	LEFT JOIN (
		SELECT ku.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage ku
			ON tc.constraint_name = ku.constraint_name
		WHERE tc.constraint_type = 'PRIMARY KEY'
			AND tc.table_name = $1
	) pk ON c.column_name = pk.column_name
	WHERE c.table_name = $1
	ORDER BY c.ordinal_position`
}

// GetIndexes method    获取PostgreSQL表的所有索引.
func (a *PostgreSQLAdapter) GetIndexes(ctx context.Context, database, table string) ([]Index, error) {
	query := `SELECT
		i.relname as index_name,
		a.attname as column_name,
		ix.indisunique as is_unique,
		ix.indisprimary as is_primary
	FROM pg_class t
	JOIN pg_index ix ON t.oid = ix.indrelid
	JOIN pg_class i ON i.oid = ix.indexrelid
	JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
	WHERE t.relname = $1
	ORDER BY i.relname, a.attnum`

	rows, err := a.db.QueryContext(ctx, query, table)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询索引信息失败: %s", err))
	}
	//nolint:errcheck
	defer rows.Close()

	indexMap := make(map[string]*Index)

	for rows.Next() {
		var (
			indexName  string
			columnName string
			isUnique   bool
			isPrimary  bool
		)

		if err := rows.Scan(&indexName, &columnName, &isUnique, &isPrimary); err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
				fmt.Sprintf("扫描索引信息失败: %s", err))
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &Index{
				Name:      indexName,
				Columns:   []string{columnName},
				IsUnique:  isUnique,
				IsPrimary: isPrimary,
			}
		}
	}

	var indexes []Index
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}

	return indexes, nil
}

// TypeMapping method    获取PostgreSQL类型映射.
func (a *PostgreSQLAdapter) TypeMapping() map[string]string {
	return map[string]string{
		"int2":        "int",
		"int4":        "int",
		"int8":        "int64",
		"smallint":    "int",
		"integer":     "int",
		"bigint":      "int64",
		"serial":      "int",
		"bigserial":   "int64",
		"varchar":     "string",
		"char":        "string",
		"text":        "string",
		"bpchar":      "string",
		"uuid":        "string",
		"bool":        "bool",
		"boolean":     "bool",
		"float4":      "float32",
		"float8":      "float64",
		"real":        "float32",
		"numeric":     "float64",
		"decimal":     "float64",
		"date":        "time.Time",
		"time":        "time.Time",
		"timestamp":   "time.Time",
		"timestamptz": "time.Time",
		"bytea":       "[]byte",
		"json":        "interface{}",
		"jsonb":       "interface{}",
	}
}

// mapPostgreSQLTypeToGo function    映射PostgreSQL类型到Go类型.
func mapPostgreSQLTypeToGo(pgType string, nullable bool, typeMap map[string]string) string {
	if nullable {
		switch pgType {
		case "int2", "int4", "int8", "smallint", "integer", "bigint", "serial", "bigserial":
			return "sql.NullInt64"
		case "varchar", "char", "text", "bpchar", "uuid":
			return "sql.NullString"
		case "bool", "boolean":
			return "sql.NullBool"
		case "date", "time", "timestamp", "timestamptz":
			return "sql.NullTime"
		case "float4", "float8", "real", "numeric", "decimal":
			return "sql.NullFloat64"
		}
	}

	if goType, exists := typeMap[pgType]; exists {
		return goType
	}

	return "interface{}"
}

// BuildGormTag method    构建PostgreSQL的GORM标签.
func (a *PostgreSQLAdapter) BuildGormTag(col Column, indexes []Index) string {
	var parts []string

	// 列名
	parts = append(parts, "column:"+col.Name)

	// 主键
	if col.IsPrimaryKey {
		parts = append(parts, "primaryKey")
	}

	// 检查是否为时间字段，添加自动时间戳标签
	if isPostgreSQLTimeColumn(col) {
		if isPostgreSQLCreateTimeColumn(col.Name) {
			parts = append(parts, "autoCreateTime")
		} else if isPostgreSQLUpdateTimeColumn(col.Name) {
			parts = append(parts, "autoUpdateTime")
		}
	}

	// 类型
	parts = append(parts, "type:"+col.Type)

	// 非空
	if !col.Nullable {
		parts = append(parts, "not null")
	}

	// 默认值（PostgreSQL特殊处理）
	if col.Default != "" {
		// 对于序列（自增）或时间字段，不添加default标签
		if len(col.Default) > 7 && col.Default[:7] == "nextval" {
			// 这是序列，不添加default标签，GORM会自动处理
		} else if !isPostgreSQLCreateTimeColumn(col.Name) && !isPostgreSQLUpdateTimeColumn(col.Name) {
			parts = append(parts, "default:"+col.Default)
		}
	}

	// 注释
	if col.Comment != "" {
		parts = append(parts, "comment:"+col.Comment)
	}

	// 处理索引
	parts = append(parts, buildPostgreSQLIndexTags(col.Name, indexes)...)

	return strings.Join(parts, ";")
}

// buildPostgreSQLIndexTags function    构建PostgreSQL索引标签.
func buildPostgreSQLIndexTags(columnName string, indexes []Index) []string {
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

// isPostgreSQLCreateTimeColumn function    判断是否为创建时间字段.
func isPostgreSQLCreateTimeColumn(columnName string) bool {
	lowerName := strings.ToLower(columnName)
	return lowerName == "created_at" ||
		lowerName == "create_time" ||
		lowerName == "createtime" ||
		lowerName == "created_time" ||
		lowerName == "createdtime" ||
		lowerName == "gmt_create" ||
		lowerName == "gmtcreate"
}

// isPostgreSQLUpdateTimeColumn function    判断是否为更新时间字段.
func isPostgreSQLUpdateTimeColumn(columnName string) bool {
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

// isPostgreSQLTimeColumn function    判断是否为时间类型字段.
func isPostgreSQLTimeColumn(col Column) bool {
	// 检查Go类型
	if strings.Contains(col.GoType, "time.Time") ||
		strings.Contains(col.GoType, "sql.NullTime") {
		return true
	}

	// 检查数据库类型
	lowerType := strings.ToLower(col.Type)
	return strings.Contains(lowerType, "timestamp") ||
		strings.Contains(lowerType, "date") ||
		strings.Contains(lowerType, "time")
}
