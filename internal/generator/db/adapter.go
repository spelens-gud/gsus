package db

// 提供数据库适配器工厂.

import (
	"context"
	"fmt"

	"github.com/spelens-gud/gsus/internal/errors"
)

// IAdapter interface    数据库适配器接口.
type IAdapter interface {
	// Connect 连接数据库.
	Connect(ctx context.Context, config *Config) error
	// Close 关闭数据库连接.
	Close() error
	// GetTables 获取所有表.
	GetTables(ctx context.Context, database string, tableFilter string) ([]Table, error)
	// GetColumns 获取表的所有列.
	GetColumns(ctx context.Context, database, table string) ([]Column, error)
	// GetIndexes 获取表的所有索引.
	GetIndexes(ctx context.Context, database, table string) ([]Index, error)
	// TypeMapping 获取数据库类型到Go类型的映射.
	TypeMapping() map[string]string
}

// Adapter struct    数据库适配器工厂.
type Adapter struct {
	adapters map[Type]func() IAdapter
}

// NewAdapter function    创建适配器工厂.
func NewAdapter() *Adapter {
	f := &Adapter{
		adapters: make(map[Type]func() IAdapter),
	}
	// 注册默认适配器
	f.Register(MySQL, func() IAdapter { return &MySQLAdapter{} })
	f.Register(PostgreSQL, func() IAdapter { return &PostgreSQLAdapter{} })
	f.Register(SQLite, func() IAdapter { return &SQLiteAdapter{} })
	f.Register(MongoDB, func() IAdapter { return &MongoDBAdapter{} })
	return f
}

// Register method    注册适配器.
func (f *Adapter) Register(dbType Type, creator func() IAdapter) {
	f.adapters[dbType] = creator
}

// Create method    创建适配器实例.
func (f *Adapter) Create(dbType Type) (IAdapter, error) {
	creator, exists := f.adapters[dbType]
	if !exists {
		return nil, errors.New(errors.ErrCodeDatabase,
			fmt.Sprintf("不支持的数据库类型: %s", dbType))
	}
	return creator(), nil
}

// defaultAdapter var    默认工厂实例.
var defaultAdapter = NewAdapter()

// GetAdapter function    获取数据库适配器.
func GetAdapter(dbType Type) (IAdapter, error) {
	return defaultAdapter.Create(dbType)
}
