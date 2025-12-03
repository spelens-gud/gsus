package db

// 提供PostgreSQL数据库适配器实现.

import (
	"context"
	"database/sql"
	"fmt"

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
