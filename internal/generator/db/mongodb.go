package db

// 提供MongoDB数据库适配器实现.

import (
	"context"
	"fmt"
	"strings"

	"github.com/spelens-gud/gsus/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBAdapter struct    MongoDB适配器.
type MongoDBAdapter struct {
	client *mongo.Client
	db     *mongo.Database
}

// Connect method    连接MongoDB数据库.
func (a *MongoDBAdapter) Connect(ctx context.Context, config *Config) error {
	var uri string

	// 根据是否有用户名密码构建不同的连接字符串
	if config.User != "" && config.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%d/%s",
			config.Host,
			config.Port,
			config.Database,
		)
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("连接MongoDB失败: %s", err))
	}

	if err = client.Ping(ctx, nil); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("MongoDB连接测试失败: %s", err))
	}

	a.client = client
	a.db = client.Database(config.Database)
	return nil
}

// Close method    关闭MongoDB连接.
func (a *MongoDBAdapter) Close() error {
	if a.client != nil {
		return a.client.Disconnect(context.Background())
	}
	return nil
}

// GetTables method    获取MongoDB所有集合.
func (a *MongoDBAdapter) GetTables(ctx context.Context, database string, tableFilter string) ([]Table, error) {
	var filter interface{}
	if tableFilter != "" {
		filter = bson.D{{Key: "name", Value: tableFilter}}
	} else {
		filter = bson.D{}
	}

	collections, err := a.db.ListCollectionNames(ctx, filter)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询集合列表失败: %s", err))
	}

	var tables []Table
	for _, name := range collections {
		tables = append(tables, Table{
			Name:    name,
			Comment: "", // MongoDB集合没有注释
		})
	}

	return tables, nil
}

// GetColumns method    获取MongoDB集合的字段（通过采样文档推断）.
func (a *MongoDBAdapter) GetColumns(ctx context.Context, database, table string) ([]Column, error) {
	collection := a.db.Collection(table)

	// 采样文档以推断字段结构
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetLimit(100))
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询集合文档失败: %s", err))
	}
	//nolint:errcheck
	defer cursor.Close(ctx)

	// 收集所有字段
	fieldMap := make(map[string]*Column)

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		for key, value := range doc {
			if _, exists := fieldMap[key]; !exists {
				column := &Column{
					Name:         key,
					Type:         inferMongoType(value),
					GoType:       mapMongoTypeToGo(value),
					Nullable:     true, // MongoDB字段默认可空
					IsPrimaryKey: key == "_id",
				}
				fieldMap[key] = column
			}
		}
	}

	// 转换为切片
	var columns []Column
	// _id字段放在最前面
	if idCol, exists := fieldMap["_id"]; exists {
		columns = append(columns, *idCol)
		delete(fieldMap, "_id")
	}

	for _, col := range fieldMap {
		columns = append(columns, *col)
	}

	return columns, nil
}

// GetIndexes method    获取MongoDB集合的索引.
func (a *MongoDBAdapter) GetIndexes(ctx context.Context, database, table string) ([]Index, error) {
	collection := a.db.Collection(table)

	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("查询索引信息失败: %s", err))
	}
	//nolint:errcheck
	defer cursor.Close(ctx)

	var indexes []Index

	for cursor.Next(ctx) {
		var indexDoc bson.M
		if err := cursor.Decode(&indexDoc); err != nil {
			continue
		}

		name, _ := indexDoc["name"].(string)
		unique, _ := indexDoc["unique"].(bool)

		// 解析索引键
		var columns []string
		if key, ok := indexDoc["key"].(bson.M); ok {
			for col := range key {
				columns = append(columns, col)
			}
		}

		indexes = append(indexes, Index{
			Name:      name,
			Columns:   columns,
			IsUnique:  unique,
			IsPrimary: name == "_id_",
		})
	}

	return indexes, nil
}

// TypeMapping method    获取MongoDB类型映射.
func (a *MongoDBAdapter) TypeMapping() map[string]string {
	return map[string]string{
		"string":    "string",
		"int":       "int",
		"int32":     "int32",
		"int64":     "int64",
		"double":    "float64",
		"bool":      "bool",
		"date":      "time.Time",
		"timestamp": "time.Time",
		"objectId":  "primitive.ObjectID",
		"array":     "[]interface{}",
		"object":    "map[string]interface{}",
		"binary":    "[]byte",
		"null":      "interface{}",
	}
}

// inferBasicMongoType function    推断基本MongoDB字段类型.
func inferBasicMongoType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32:
		return "int32"
	case int64:
		return "int64"
	case float32, float64:
		return "double"
	case bool:
		return "bool"
	default:
		return ""
	}
}

// inferComplexMongoType function    推断复杂MongoDB字段类型.
func inferComplexMongoType(value interface{}) string {
	switch value.(type) {
	case bson.M:
		return "object"
	case []interface{}:
		return "array"
	case []byte:
		return "binary"
	case nil:
		return "null"
	default:
		return "interface{}"
	}
}

// inferMongoType function    推断MongoDB字段类型.
func inferMongoType(value interface{}) string {
	// 先尝试推断基本类型
	if basicType := inferBasicMongoType(value); basicType != "" {
		return basicType
	}

	// 再尝试推断复杂类型
	return inferComplexMongoType(value)
}

// mapMongoTypeToGo function    映射MongoDB类型到Go类型.
func mapMongoTypeToGo(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32:
		return "int32"
	case int64:
		return "int64"
	case float32, float64:
		return "float64"
	case bool:
		return "bool"
	case bson.M:
		return "map[string]interface{}"
	case []interface{}:
		return "[]interface{}"
	case []byte:
		return "[]byte"
	default:
		return "interface{}"
	}
}

// BuildGormTag method    构建MongoDB的GORM标签（MongoDB主要使用bson标签）.
func (a *MongoDBAdapter) BuildGormTag(col Column, indexes []Index) string {
	var parts []string

	// MongoDB使用bson标签，但为了兼容性也提供gorm标签
	// 列名（在MongoDB中是字段名）
	parts = append(parts, "column:"+col.Name)

	// 主键（MongoDB的_id字段）
	if col.IsPrimaryKey || col.Name == "_id" {
		parts = append(parts, "primaryKey")
	}

	// 检查是否为时间字段，添加自动时间戳标签
	if isMongoDBTimeColumn(col) {
		if isMongoDBCreateTimeColumn(col.Name) {
			parts = append(parts, "autoCreateTime")
		} else if isMongoDBUpdateTimeColumn(col.Name) {
			parts = append(parts, "autoUpdateTime")
		}
	}

	// 类型
	if col.Type != "" {
		parts = append(parts, "type:"+col.Type)
	}

	// MongoDB通常不强制非空，但可以标记
	if !col.Nullable {
		parts = append(parts, "not null")
	}

	// 注释
	if col.Comment != "" {
		parts = append(parts, "comment:"+col.Comment)
	}

	// MongoDB的索引处理
	parts = append(parts, buildMongoDBIndexTags(col.Name, indexes)...)

	return strings.Join(parts, ";")
}

// buildMongoDBIndexTags function    构建MongoDB索引标签.
func buildMongoDBIndexTags(columnName string, indexes []Index) []string {
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

// isMongoDBCreateTimeColumn function    判断是否为创建时间字段.
func isMongoDBCreateTimeColumn(columnName string) bool {
	lowerName := strings.ToLower(columnName)
	return lowerName == "created_at" ||
		lowerName == "createdat" ||
		lowerName == "create_time" ||
		lowerName == "createtime" ||
		lowerName == "created_time" ||
		lowerName == "createdtime" ||
		lowerName == "gmt_create" ||
		lowerName == "gmtcreate"
}

// isMongoDBUpdateTimeColumn function    判断是否为更新时间字段.
func isMongoDBUpdateTimeColumn(columnName string) bool {
	lowerName := strings.ToLower(columnName)
	return lowerName == "updated_at" ||
		lowerName == "updatedat" ||
		lowerName == "update_time" ||
		lowerName == "updatetime" ||
		lowerName == "updated_time" ||
		lowerName == "updatedtime" ||
		lowerName == "gmt_modified" ||
		lowerName == "gmtmodified" ||
		lowerName == "modified_at" ||
		lowerName == "modify_time"
}

// isMongoDBTimeColumn function    判断是否为时间类型字段.
func isMongoDBTimeColumn(col Column) bool {
	// 检查Go类型
	if strings.Contains(col.GoType, "time.Time") ||
		strings.Contains(col.GoType, "primitive.DateTime") {
		return true
	}

	// 检查数据库类型
	lowerType := strings.ToLower(col.Type)
	return lowerType == "date" ||
		lowerType == "timestamp" ||
		strings.Contains(lowerType, "datetime")
}
