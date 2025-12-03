package db

// 提供MongoDB数据库适配器实现.

import (
	"context"
	"fmt"
	"strings"

	"github.com/spelens-gud/gsus/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		a.processDocumentFields(doc, fieldMap)
	}

	// 转换为切片
	columns := make([]Column, 0)
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

// processDocumentFields function    处理文档字段.
func (a *MongoDBAdapter) processDocumentFields(doc bson.M, fieldMap map[string]*Column) {
	for key, value := range doc {
		if _, exists := fieldMap[key]; !exists {
			goType, dbType, isPK := a.extractFieldInfo(key, value)

			column := &Column{
				Name:         key,
				Type:         dbType,
				GoType:       goType,
				Nullable:     true, // MongoDB字段默认可空
				IsPrimaryKey: isPK,
			}
			fieldMap[key] = column
		}
	}
}

// extractFieldInfo function    提取字段信息.
func (a *MongoDBAdapter) extractFieldInfo(key string, value interface{}) (goType, dbType string, isPK bool) {
	goType = mapMongoTypeToGo(value)
	dbType = inferMongoType(value)
	isPK = false

	// 特殊处理 _id 字段
	if key == "_id" {
		goType = "primitive.ObjectID"
		dbType = "objectId"
		isPK = true
	}

	// 特殊处理时间字段
	if isMongoDBCreateTimeColumn(key) || isMongoDBUpdateTimeColumn(key) {
		goType = "time.Time"
		dbType = "timestamp"
	}

	return goType, dbType, isPK
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
		"array":     "bson.A",
		"object":    "bson.M",
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
	case []interface{}, bson.A:
		return "array"
	case []byte:
		return "binary"
	case nil:
		return "null"
	case primitive.ObjectID:
		return "objectId"
	case primitive.DateTime:
		return "timestamp"
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
	if goType, ok := isBasicType(value); ok {
		return goType
	}

	if goType, ok := isComplexType(value); ok {
		return goType
	}

	return "interface{}"
}

// BuildGormTag method    构建MongoDB的GORM标签（返回gorm标签内容，不包含gorm:前缀）.
func (a *MongoDBAdapter) BuildGormTag(col Column, indexes []Index) string {
	var parts []string

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

	// 注释
	if col.Comment != "" {
		parts = append(parts, "comment:"+col.Comment)
	}

	// MongoDB的索引处理
	parts = append(parts, buildMongoDBIndexTags(col.Name, indexes)...)

	return strings.Join(parts, ";")
}

// BuildMongoDBTags method    构建MongoDB的完整标签（包括gorm、bson、json）.
func (a *MongoDBAdapter) BuildMongoDBTags(col Column, indexes []Index) string {
	var tags []string

	// GORM 标签
	gormTag := a.BuildGormTag(col, indexes)
	if gormTag != "" {
		tags = append(tags, fmt.Sprintf(`gorm:"%s"`, gormTag))
	}

	// BSON 标签（MongoDB 原生标签）
	var bsonParts []string
	bsonParts = append(bsonParts, col.Name)
	// _id 字段不应该有 omitempty，因为它是 MongoDB 的必需字段
	// 其他字段根据 Nullable 决定是否添加 omitempty
	if col.Nullable && col.Name != "_id" {
		bsonParts = append(bsonParts, "omitempty")
	}
	tags = append(tags, fmt.Sprintf(`bson:"%s"`, strings.Join(bsonParts, ",")))

	// JSON 标签
	var jsonParts []string
	jsonParts = append(jsonParts, col.Name)
	// JSON 标签也添加 omitempty（除了 _id）
	if col.Nullable && col.Name != "_id" {
		jsonParts = append(jsonParts, "omitempty")
	}
	tags = append(tags, fmt.Sprintf(`json:"%s"`, strings.Join(jsonParts, ",")))

	return strings.Join(tags, " ")
}

// buildMongoDBIndexTags function    构建MongoDB索引标签.
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

// isMongoDBCreateTimeColumn function    判断是否为创建时间字段.
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

// isMongoDBUpdateTimeColumn function    判断是否为更新时间字段.
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

// isMongoDBTimeColumn function    判断是否为时间类型字段.
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

// isBasicType function    检查是否为基本类型.
func isBasicType(value interface{}) (string, bool) {
	switch value.(type) {
	case string:
		return "string", true
	case int, int32:
		return "int32", true
	case int64:
		return "int64", true
	case float32, float64:
		return "float64", true
	case bool:
		return "bool", true
	default:
		return "", false
	}
}

// isComplexType function    检查是否为复杂类型.
func isComplexType(value interface{}) (string, bool) {
	switch value.(type) {
	case bson.M:
		return "bson.M", true
	case bson.A, []interface{}:
		return "bson.A", true
	case []byte:
		return "[]byte", true
	case bson.D:
		return "bson.D", true
	case primitive.ObjectID:
		return "primitive.ObjectID", true
	case primitive.DateTime:
		return "time.Time", true
	default:
		return "", false
	}
}
