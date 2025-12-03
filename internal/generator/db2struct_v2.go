// Package generator 提供数据库表结构生成功能.
package generator

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/generator/db"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
)

// Generator struct    数据库结构生成器.
type Generator[T db.IAdapter] struct {
	adapter T
	config  *db.Config
	opts    *config.DbOpt
}

// NewGenerator function    创建生成器实例.
func NewGenerator[T db.IAdapter](adapter T, cfg *db.Config, opts *config.DbOpt) *Generator[T] {
	return &Generator[T]{
		adapter: adapter,
		config:  cfg,
		opts:    opts,
	}
}

// GenerateAll method    生成所有表结构.
func (g *Generator[T]) GenerateAll(ctx context.Context, outputDir string) error {
	// 连接数据库
	if err := g.adapter.Connect(ctx, g.config); err != nil {
		return err
	}
	//nolint:errcheck
	defer g.adapter.Close()

	// 获取所有表
	tables, err := g.adapter.GetTables(ctx, g.config.Database, "")
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("获取表列表失败: %s", err))
	}

	// 生成每个表的结构
	for _, table := range tables {
		if err := g.GenerateTable(ctx, table.Name, outputDir); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeGenerate,
				fmt.Sprintf("生成表 %s 失败: %s", table.Name, err))
		}
	}

	return nil
}

// GenerateTable method    生成单个表结构.
func (g *Generator[T]) GenerateTable(ctx context.Context, tableName, outputDir string) error {
	// 获取表信息
	tables, err := g.adapter.GetTables(ctx, g.config.Database, tableName)
	if err != nil || len(tables) == 0 {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("获取表 %s 信息失败", tableName))
	}
	table := tables[0]

	// 获取列信息
	columns, err := g.adapter.GetColumns(ctx, g.config.Database, tableName)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("获取表 %s 列信息失败: %s", tableName, err))
	}

	// 获取索引信息
	indexes, err := g.adapter.GetIndexes(ctx, g.config.Database, tableName)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase,
			fmt.Sprintf("获取表 %s 索引信息失败: %s", tableName, err))
	}

	// 生成结构体代码
	code, err := g.generateStructCode(table, columns, indexes)
	if err != nil {
		return err
	}

	// 写入文件
	filename := filepath.Join(outputDir, strcase.SnakeCase(tableName)+".go")
	if err := utils.ImportAndWrite(code, filename); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile,
			fmt.Sprintf("写入文件失败: %s", err))
	}

	return nil
}

// generateStructCode method    生成结构体代码.
func (g *Generator[T]) generateStructCode(table db.Table, columns []db.Column, indexes []db.Index) ([]byte, error) {
	structName := strcase.UpperCamelCase(table.Name)

	// 生成字段定义
	fields, fieldNameMap := g.generateFields(columns, indexes)

	// 生成基础结构
	src := fmt.Sprintf(template.HeadTemplate,
		table.Comment,
		table.Name,
		g.opts.PkgName,
		table.Name,
		": "+table.Comment,
		structName,
		fields,
	)

	// 生成TableName方法
	tableNameFunc := fmt.Sprintf(template.TableFuncTemplate,
		structName,
		table.Name,
		utils.GetFuncCallerIdent(structName),
		structName,
		structName,
	)

	data := []byte(src + "\n\n" + tableNameFunc)

	// 处理标签
	data, err := g.processTags(table.Name, fieldNameMap, columns, indexes, data)
	if err != nil {
		return nil, err
	}

	// 生成泛型方法
	if len(g.opts.GenericOption) > 0 {
		genericF, err := parser.NewType(structName, g.opts.PkgName, g.opts.GenericOption...)
		if err != nil {
			return nil, errors.WrapWithCode(err, errors.ErrCodeGenerate,
				fmt.Sprintf("生成泛型方法失败: %s", err))
		}
		data = append(data, genericF...)
	}

	return data, nil
}

// generateFields method    生成字段定义.
func (g *Generator[T]) generateFields(columns []db.Column, indexes []db.Index) (string, map[string]string) {
	var builder strings.Builder
	builder.WriteString("struct {")

	fieldNameMap := make(map[string]string)

	for _, col := range columns {
		goType := col.GoType
		if rp := g.opts.TypeReplace[goType]; len(rp) > 0 {
			goType = rp
		}

		fieldName := fmtFieldName(stringifyFirstChar(col.Name))
		fieldNameMap[fieldName] = col.Name

		// 生成字段
		builder.WriteString(fmt.Sprintf("\n%s %s", fieldName, goType))

		// 生成标签
		var tags []string
		if g.opts.GormAnnotation {
			// 使用适配器的BuildGormTag方法
			tag := g.adapter.BuildGormTag(col, indexes)
			if tag != "" {
				tags = append(tags, fmt.Sprintf(`gorm:"%s"`, tag))
			}
		}

		if len(tags) > 0 {
			builder.WriteString(fmt.Sprintf(" `%s`", strings.Join(tags, " ")))
		}

		// 添加注释
		if (g.opts.CommentOutside || !g.opts.SqlInfo) && col.Comment != "" {
			builder.WriteString(" // " + col.Comment)
		}
	}

	builder.WriteString("\n}")
	return builder.String(), fieldNameMap
}

// buildGormTag method    构建GORM标签.
// Deprecated: This package is deprecated and will be removed in future versions.
//
//nolint:unused
func (g *Generator[T]) buildGormTag(col db.Column, indexes []db.Index) string {
	var parts []string

	parts = append(parts, "column:"+col.Name)

	if col.IsPrimaryKey {
		parts = append(parts, "primary_key")
	}

	if col.IsAutoIncr {
		parts = append(parts, "auto_increment")
	}

	parts = append(parts, "type:"+col.Type)

	if !col.Nullable {
		parts = append(parts, "not null")
	}

	// 处理索引
	indexParts := g.buildIndexTags(col.Name, indexes)
	parts = append(parts, indexParts...)

	return strings.Join(parts, ";")
}

// buildIndexTags method    构建索引标签.
//
//nolint:unused
func (g *Generator[T]) buildIndexTags(columnName string, indexes []db.Index) []string {
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
		parts = append(parts, "unique_index:"+strings.Join(uniqueIndexes, ","))
	}

	return parts
}

// processTags method    处理标签.
func (g *Generator[T]) processTags(tableName string, fieldNameMap map[string]string,
	columns []db.Column, indexes []db.Index, data []byte) ([]byte, error) {

	columnMap := make(map[string]db.Column)
	for _, col := range columns {
		fieldName := fmtFieldName(stringifyFirstChar(col.Name))
		columnMap[fieldName] = col
	}

	var tagOptions []TagOption

	// GORM/SQL标签
	gormTagOpt, sqlTagOpt := g.genTagUpdateFunc(tableName, fieldNameMap, columnMap, indexes)
	tagOptions = append(tagOptions, gormTagOpt)
	if g.opts.SqlInfo {
		tagOptions = append(tagOptions, sqlTagOpt)
	}

	// JSON标签
	switch g.opts.Json {
	case "camel":
		tagOptions = append(tagOptions, CamelCase("json", false))
	case "snake":
		tagOptions = append(tagOptions, SnakeCase("json", false))
	}

	result, _, err := ParseInput(data, tagOptions...)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeParse,
			fmt.Sprintf("处理标签失败: %s", err))
	}

	return result, nil
}

// createGormTagFunc method    创建GORM标签函数.
func (g *Generator[T]) createGormTagFunc(columnMap map[string]db.Column, fieldNameMap map[string]string,
	indexMap map[string][]db.Index) TagFunc {
	return func(structName, fieldName, newTag, oldTag string) string {
		col, ok := columnMap[fieldName]
		if !ok {
			return oldTag
		}
		return buildGormTagProperties(col, fieldNameMap, indexMap, fieldName, g.opts.GormAnnotation)
	}
}

// createSqlTagFunc method    创建SQL标签函数.
func (g *Generator[T]) createSqlTagFunc(columnMap map[string]db.Column) TagFunc {
	return func(structName, fieldName, newTag, oldTag string) string {
		col, ok := columnMap[fieldName]
		if !ok {
			return ""
		}
		return buildSqlTagProperties(col)
	}
}

// createCombinedSqlTagFunc method    创建组合SQL标签函数.
func (g *Generator[T]) createCombinedSqlTagFunc(sqlTagFunc TagFunc) TagFunc {
	return func(structName, fieldName, newTag, oldTag string) string {
		sqlTag := sqlTagFunc(structName, fieldName, newTag, oldTag)
		if sqlTag == "" {
			return oldTag
		}
		if oldTag == "" {
			return sqlTag
		}
		return strings.Join([]string{oldTag, sqlTag}, ";")
	}
}

// buildTagOptions method    构建标签选项.
func (g *Generator[T]) buildTagOptions(gormTagFunc, sqlTagFunc TagFunc) (TagOption, TagOption) {
	if g.opts.SqlTag == "sql" {
		return TagOption{
				Tag:        g.opts.SqlTag,
				Type:       TypeSnakeCase,
				Cover:      true,
				Edit:       true,
				AppendFunc: gormTagFunc,
			}, TagOption{
				Tag:        "sql",
				Type:       TypeSnakeCase,
				Cover:      true,
				Edit:       true,
				AppendFunc: g.createCombinedSqlTagFunc(sqlTagFunc),
			}
	}

	return TagOption{
			Tag:        g.opts.SqlTag,
			Type:       TypeSnakeCase,
			Cover:      true,
			Edit:       true,
			AppendFunc: gormTagFunc,
		}, TagOption{
			Tag:        "sql",
			Type:       TypeSnakeCase,
			Cover:      true,
			Edit:       true,
			AppendFunc: sqlTagFunc,
		}
}

// genTagUpdateFunc method    生成标签更新函数.
func (g *Generator[T]) genTagUpdateFunc(tableName string, fieldNameMap map[string]string,
	columnMap map[string]db.Column, indexes []db.Index) (gormTag TagOption, sqlTag TagOption) {

	// 构建索引映射
	indexMap := buildIndexMap(indexes)

	// 处理注释外置选项
	clearColumnComments(columnMap, g.opts.CommentOutside)

	// 创建标签函数
	gormTagFunc := g.createGormTagFunc(columnMap, fieldNameMap, indexMap)
	sqlTagFunc := g.createSqlTagFunc(columnMap)

	// 构建标签选项
	return g.buildTagOptions(gormTagFunc, sqlTagFunc)
}

// buildIndexMap function    构建字段到索引的映射关系.
func buildIndexMap(indexes []db.Index) map[string][]db.Index {
	indexMap := make(map[string][]db.Index)
	for _, idx := range indexes {
		for _, col := range idx.Columns {
			fieldName := fmtFieldName(stringifyFirstChar(col))
			indexMap[fieldName] = append(indexMap[fieldName], idx)
		}
	}
	return indexMap
}

// clearColumnComments function    处理注释外置选项.
func clearColumnComments(columnMap map[string]db.Column, commentOutside bool) {
	if commentOutside {
		for k, v := range columnMap {
			v.Comment = ""
			columnMap[k] = v
		}
	}
}

// buildGormTagProperties function    构建GORM标签属性.
func buildGormTagProperties(col db.Column, fieldNameMap map[string]string, indexMap map[string][]db.Index,
	fieldName string, gormAnnotation bool) string {
	var props []string

	if gormAnnotation {
		props = append(props, "column:"+fieldNameMap[fieldName])
	}

	if col.IsPrimaryKey {
		props = append(props, "primary_key")
	}

	if col.IsAutoIncr {
		props = append(props, "autoIncrement")
	}

	props = append(props, "type:"+col.Type)

	if !col.Nullable {
		props = append(props, "not null")
	}

	// 处理索引
	props = append(props, buildIndexProperties(indexMap, fieldName)...)

	return strings.Join(props, ";")
}

// buildIndexProperties function    构建索引属性.
func buildIndexProperties(indexMap map[string][]db.Index, fieldName string) []string {
	idxs, exists := indexMap[fieldName]
	if !exists {
		return nil
	}

	return buildIndexLists(idxs)
}

// buildIndexLists function    构建索引列表.
func buildIndexLists(idxs []db.Index) []string {
	var (
		props         []string
		normalIndexes []string
		uniqueIndexes []string
	)

	for _, idx := range idxs {
		if idx.IsPrimary {
			continue
		}

		if idx.IsUnique {
			uniqueIndexes = append(uniqueIndexes, idx.Name)
		} else {
			normalIndexes = append(normalIndexes, idx.Name)
		}
	}

	if len(normalIndexes) > 0 {
		props = append(props, "index:"+strings.Join(normalIndexes, ","))
	}

	if len(uniqueIndexes) > 0 {
		props = append(props, "unique_index:"+strings.Join(uniqueIndexes, ","))
	}

	return props
}

// buildSqlTagProperties function    构建SQL标签属性.
func buildSqlTagProperties(col db.Column) string {
	var props []string

	if col.Comment != "" {
		props = append(props, "comment:'"+col.Comment+"'")
	}

	if col.Default != "" {
		d := "default:" + col.Default
		if col.Extra != "" {
			d += " " + col.Extra
		}
		props = append(props, d)
	}

	return strings.Join(props, ";")
}

// GenAllDb2StructWithAdapter function    使用适配器生成所有表结构.
func GenAllDb2StructWithAdapter(ctx context.Context, dir string, dbConfig *db.Config,
	options ...config.DbOption) error {

	opts := config.NewDbOpt(options...)

	// 获取适配器
	adapter, err := db.GetAdapter(dbConfig.Type)
	if err != nil {
		return err
	}

	// 创建生成器
	gen := &Generator[db.IAdapter]{
		adapter: adapter,
		config:  dbConfig,
		opts:    opts,
	}

	return gen.GenerateAll(ctx, dir)
}

// GenTableWithAdapter function    使用适配器生成单个表结构.
func GenTableWithAdapter(ctx context.Context, tableName string, dbConfig *db.Config,
	options ...config.DbOption) ([]byte, error) {

	opts := config.NewDbOpt(options...)

	// 获取适配器
	adapter, err := db.GetAdapter(dbConfig.Type)
	if err != nil {
		return nil, err
	}

	// 连接数据库
	if err := adapter.Connect(ctx, dbConfig); err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer adapter.Close()

	// 获取表信息
	tables, err := adapter.GetTables(ctx, dbConfig.Database, tableName)
	if err != nil || len(tables) == 0 {
		return nil, errors.New(errors.ErrCodeDatabase,
			fmt.Sprintf("获取表 %s 信息失败", tableName))
	}

	// 获取列信息
	columns, err := adapter.GetColumns(ctx, dbConfig.Database, tableName)
	if err != nil {
		return nil, err
	}

	// 获取索引信息
	indexes, err := adapter.GetIndexes(ctx, dbConfig.Database, tableName)
	if err != nil {
		return nil, err
	}

	// 创建生成器
	gen := &Generator[db.IAdapter]{
		adapter: adapter,
		config:  dbConfig,
		opts:    opts,
	}

	return gen.generateStructCode(tables[0], columns, indexes)
}
