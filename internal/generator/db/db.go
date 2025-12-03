// Package db 包级别数据库适配器, 使用数据库的具体生成操作
package db

// 提供数据库适配器接口和通用类型定义.

// Type    数据库类型.
type Type string

const (
	// MySQL 数据库类型.
	MySQL Type = "mysql"
	// PostgreSQL 数据库类型.
	PostgreSQL Type = "postgresql"
	// SQLite 数据库类型.
	SQLite Type = "sqlite"
	// MongoDB 数据库类型.
	MongoDB Type = "mongodb"
)

// Config struct    数据库连接配置.
type Config struct {
	Type     Type              // 数据库类型
	User     string            // 数据库用户名
	Password string            // 密码
	Host     string            // 数据库主机地址
	Port     int               // 数据库端口
	Database string            // 数据库名称
	Charset  string            // 字符集
	Extra    map[string]string // 额外的连接参数
}

// Table struct    表信息.
type Table struct {
	Name    string // 表名
	Comment string // 表注释
	Schema  string // 模式名（PostgreSQL）
}

// Column struct    列信息.
type Column struct {
	Name         string // 列名
	Type         string // 数据库类型
	GoType       string // Go类型
	Nullable     bool   // 是否可空
	IsPrimaryKey bool   // 是否主键
	IsAutoIncr   bool   // 是否自增
	Default      string // 默认值
	Comment      string // 注释
	Extra        string // 额外信息
}

// Index struct    索引信息.
type Index struct {
	Name       string   // 索引名称
	Columns    []string // 索引列
	IsUnique   bool     // 是否唯一索引
	IsPrimary  bool     // 是否主键索引
	IsFullText bool     // 是否全文索引
}
