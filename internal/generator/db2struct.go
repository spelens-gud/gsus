package generator

// Deprecated: This package is deprecated and will be removed in future versions.
// Please use the new package instead.

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/spelens-gud/gsus/internal/parser"
	"github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
	"github.com/stoewer/go-strcase"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	golangByteArray = "[]byte"
	golangInt       = "int"
	golangInt64     = "int64"
	golangFloat32   = "float32"
	golangFloat64   = "float64"
	golangTime      = "time.Time"
	golangInterface = "interface{}"
	sqlNullInt      = "sql.NullInt64"
	sqlNullFloat    = "sql.NullFloat64"
	sqlNullString   = "sql.NullString"
	sqlNullTime     = "sql.NullTime"
)

var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
}

var intToWordMap = []string{
	"zero",
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
}

// 定义MySQL类型到Go类型的映射.
var mysqlToGoTypeMap = map[string]string{
	"tinyint":    golangInt,
	"int":        golangInt,
	"smallint":   golangInt,
	"mediumint":  golangInt,
	"bigint":     golangInt64,
	"char":       "string",
	"enum":       "string",
	"varchar":    "string",
	"longtext":   "string",
	"mediumtext": "string",
	"text":       "string",
	"tinytext":   "string",
	"date":       golangTime,
	"datetime":   golangTime,
	"time":       golangTime,
	"timestamp":  golangTime,
	"decimal":    golangFloat64,
	"double":     golangFloat64,
	"float":      golangFloat32,
	"binary":     golangByteArray,
	"blob":       golangByteArray,
	"longblob":   golangByteArray,
	"mediumblob": golangByteArray,
	"varbinary":  golangByteArray,
	"json":       golangInterface,
}

// Table struct    表结构体.
type Table struct {
	Name    string // 表名
	Comment string // 表注释
}

// DBConfig struct    数据库连接信息.
type DBConfig struct {
	User     string // 数据库用户名
	Password string // 密码
	DB       string // 数据库名称
	Host     string // 数据库主机地址
	Port     int    // 数据库端口
	Cstr     string // 数据库连接字符串
}

// TagFunc    tag处理函数.
type TagFunc func(structName, fieldName, newTag, oldTag string) (tag string)

// Deprecated: This package is deprecated and will be removed in future versions.
// GenAllDb2Struct    生成所有表结构.
func GenAllDb2Struct(dir string, dbConfig DBConfig, options ...config.DbOption) (err error) {
	db, err := parser.GetConnection(dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB,
	)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase, fmt.Sprintf("连接数据库失败: %s", err))
	}

	tables, err := getTables(db, dbConfig.DB, "")
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeDatabase, fmt.Sprintf("获取表失败: %s", err))
	}

	for _, table := range tables {
		var ret []byte
		ret, err = GenTable(table.Name, dbConfig, options...)
		if err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成表结构失败: %s", err))
		}

		if err = utils.ImportAndWrite(ret, filepath.Join(dir, strcase.SnakeCase(table.Name)+".go")); err != nil {
			return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("写入文件失败: %s", err))
		}
	}
	return nil
}

// getTables    获取数据库所有表.
// tableName为空的时候获取所有表,否则获取tableName的表.
func getTables(db *sql.DB, dbName string, tableName string) (ts []Table, err error) {
	sqlCommand := `SELECT TABLE_NAME, TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ?`
	args := []interface{}{dbName}
	if len(tableName) > 0 {
		sqlCommand += " AND TABLE_NAME = ?"
		args = append(args, tableName)
	}
	r, err := db.Query(sqlCommand, args...)
	if err != nil {
		return nil, err
	}
	// nolint
	defer r.Close()
	for r.Next() {
		var table Table
		if err = r.Scan(&table.Name, &table.Comment); err != nil {
			return nil, err
		}
		ts = append(ts, table)
	}
	return ts, nil
}

// GenTable    生成表结构.
func GenTable(table string, dbConfig DBConfig, options ...config.DbOption) (data []byte, err error) {
	// 原db2struct返回的顺序是字段字典序的 已改造成 db2structx
	db, err := parser.GetConnection(dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase, fmt.Sprintf("连接数据库失败: %s", err))
	}
	// nolint
	defer db.Close()

	// 获取表结构
	res, err := parser.GetColumnsFromMysqlTable(db, dbConfig.DB, table)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase, fmt.Sprintf("获取表结构失败: %s", err))
	}

	// 获取表名
	tables, err := getTables(db, dbConfig.DB, table)
	if err != nil || len(tables) != 1 {
		return nil, errors.WrapWithCode(err, errors.ErrCodeDatabase, fmt.Sprintf("获取表结构失败: %s", err))
	}

	var (
		opts       = config.NewDbOpt(options...) // 配置
		structName = strcase.SnakeCase(table)    // 默认结构体名称
	)

	// 基础结构生成
	data, fieldNameMap := GenerateStruct(*res, table, structName, opts.PkgName, tables[0].Comment, opts)

	// 结构体tag
	if data, err = parseTag(table, fieldNameMap, dbConfig, data, opts); err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成表结构失败: %s", err))
	}

	// 泛型方法
	genericF, err := parser.NewType(structName, opts.PkgName, opts.GenericOption...)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成表结构失败: %s", err))
	}

	data = append(data, genericF...)
	return data, nil
}

// GenerateStruct    生成结构体.
func GenerateStruct(columnTypes []parser.CTypes, tableName, structName, pkgName,
	tableComment string, opts *config.DbOpt) ([]byte, map[string]string) {

	// 生成mysql专用结构体
	dbTypes, fieldNameMap := generateMysqlTypes(columnTypes, 0, opts)
	src := fmt.Sprintf(template.HeadTemplate, tableComment, tableName, pkgName, tableName,
		": "+tableComment, structName, dbTypes)

	tableNameFunc := fmt.Sprintf(template.TableFuncTemplate, structName, tableName,
		utils.GetFuncCallerIdent(structName), structName, structName)
	return []byte(src + "\n\n" + tableNameFunc), fieldNameMap
}

// generateMysqlTypes    生成mysql类型.
func generateMysqlTypes(objs []parser.CTypes, depth int, opts *config.DbOpt) (string, map[string]string) {
	structure := "struct {"
	fieldNameMap := make(map[string]string)
	for _, obj := range objs {
		mysqlType := obj.Info
		nullable := false
		if mysqlType["nullable"] == "YES" {
			nullable = true
		}

		primary := ""
		if mysqlType["primary"] == "PRI" {
			primary = ";primary_key"
		}

		var valueType string

		valueType = mysqlTypeToGoType(mysqlType["value"], nullable)
		if rp := opts.TypeReplace[valueType]; len(rp) > 0 {
			valueType = rp
		}
		fieldName := fmtFieldName(stringifyFirstChar(obj.Key))
		fieldNameMap[fieldName] = obj.Key
		var annotations []string
		if opts.GormAnnotation {
			annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s%s\"", obj.Key, primary))
		}
		if len(annotations) > 0 {
			structure += fmt.Sprintf("\n%s %s `%s`",
				fieldName,
				valueType,
				strings.Join(annotations, " "))
		} else {
			structure += fmt.Sprintf("\n%s %s",
				fieldName,
				valueType)
		}
		if (opts.CommentOutside || !opts.SqlInfo) && len(obj.Info["comment"]) > 0 {
			structure += " // " + obj.Info["comment"]
		}
	}
	structure += "\n}"
	return structure, fieldNameMap
}

// mysqlTypeToGoType    将mysql类型转换为go类型.
func mysqlTypeToGoType(mysqlType string, nullable bool) string {
	// 处理可空类型
	if nullable {
		switch mysqlType {
		case "tinyint", "int", "smallint", "mediumint", "bigint":
			return sqlNullInt
		case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
			return sqlNullString
		case "date", "datetime", "time", "timestamp":
			return sqlNullTime
		case "decimal", "double", "float":
			return sqlNullFloat
		}
	}

	// 查找映射表
	if goType, exists := mysqlToGoTypeMap[mysqlType]; exists {
		return goType
	}

	return golangInterface
}

// fmtFieldName    格式化字段名.
func fmtFieldName(s string) string {
	// 全大写的话 先转成小写
	if strings.ToUpper(s) == s {
		s = strings.ToLower(s)
	}
	name := lintFieldName(s)
	runes := []rune(name)
	for i, c := range runes {
		ok := unicode.IsLetter(c) || unicode.IsDigit(c)
		if i == 0 {
			ok = unicode.IsLetter(c)
		}
		if !ok {
			runes[i] = '_'
		}
	}
	return string(runes)
}

// lintFieldName    字段名格式化.
func lintFieldName(name string) string {
	if isSimpleCase(name) {
		return name
	}

	// 移除开头的下划线
	name = removeLeadingUnderscores(name)

	// 检查是否全小写
	if isAllLowerCase(name) {
		return processAllLowerCase(name)
	}

	// 处理驼峰命名和下划线混合的情况
	return processCamelCaseAndUnderscore(name)
}

// processCamelCaseAndUnderscore    处理驼峰命名和下划线混合的情况.
func processCamelCaseAndUnderscore(name string) string {
	runes := []rune(name)
	w, i := 0, 0
	for i+1 <= len(runes) {
		eow, n := checkEndOfWordCondition(runes, i)

		if !eow {
			i++
			continue
		}

		if i+1 < len(runes) && runes[i+1] == '_' {
			runes = processUnderscoreSequence(runes, i, n)
		}

		i++
		w = handleEndOfWord(runes, i, w)
	}
	return string(runes)
}

// checkEndOfWordCondition    检查是否到达单词结尾.
func checkEndOfWordCondition(runes []rune, i int) (bool, int) {
	if i+1 == len(runes) {
		return true, 0
	} else if runes[i+1] == '_' {
		n := 1
		for i+n+1 < len(runes) && runes[i+n+1] == '_' {
			n++
		}
		return true, n
	} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
		return true, 0
	}
	return false, 0
}

// processUnderscoreSequence    处理下划线序列.
func processUnderscoreSequence(runes []rune, i, n int) []rune {
	if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
		n--
	}

	copy(runes[i+1:], runes[i+n+1:])
	return runes[:len(runes)-n]
}

// handleEndOfWord    处理单词结束.
func handleEndOfWord(runes []rune, i, w int) int {
	word := string(runes[w:i])
	if u := strings.ToUpper(word); commonInitialisms[u] {
		copy(runes[w:], []rune(u))
	} else if strings.ToLower(word) == word {
		runes[w] = unicode.ToUpper(runes[w])
	}
	return i
}

// isSimpleCase    检查简单情况.
func isSimpleCase(name string) bool {
	return name == "_"
}

// removeLeadingUnderscores    移除开头的下划线.
func removeLeadingUnderscores(name string) string {
	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}
	return name
}

// isAllLowerCase    检查是否全小写.
func isAllLowerCase(name string) bool {
	for _, r := range name {
		if !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// processAllLowerCase    处理全小写情况.
func processAllLowerCase(name string) string {
	runes := []rune(name)
	if u := strings.ToUpper(name); commonInitialisms[u] {
		copy(runes[0:], []rune(u))
	} else {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}

// stringifyFirstChar    将字符串首字母转为大写.
func stringifyFirstChar(str string) string {
	first := str[:1]

	i, err := strconv.ParseInt(first, 10, 8)

	if err != nil {
		return str
	}

	return intToWordMap[i] + "_" + str[1:]
}

// parseTag    解析标签.
func parseTag(table string, fieldNameMap map[string]string, dbConfig DBConfig,
	in []byte, opts *config.DbOpt) (data []byte, err error) {

	// 生成标签更新函数
	gto, sto := GenTagUpdateFuncByTable(table, fieldNameMap, dbConfig, opts)
	pos := []TagOption{gto}
	if opts.SqlInfo {
		pos = append(pos, sto)
	}
	switch opts.Json {
	case "camel":
		{
			pos = append(pos, CamelCase("json", false))
		}
	case "snake":
		{
			pos = append(pos, SnakeCase("json", false))
		}
	}
	data, _, err = ParseInput(in, pos...)
	if err != nil {
		return nil, errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("格式化源代码失败: %s", err))
	}
	return data, nil
}

// GenTagUpdateFuncByTable    生成表结构体标签更新函数.
func GenTagUpdateFuncByTable(table string, fieldNameMap map[string]string,
	dbConfig DBConfig, opts *config.DbOpt) (gorm TagOption, sql TagOption) {

	// 初始化数据库连接
	db := initConnect(&dbConfig)
	// 获取表结构
	m, idm := initTableDesc(table, fieldNameMap, db)
	if opts.CommentOutside {
		for k, v := range m {
			v.ColumnComment = ""
			m[k] = v
		}
	}
	if opts.SqlTag == "sql" {
		return TagOption{
				Tag:        opts.SqlTag,
				Type:       TypeSnakeCase,
				Cover:      true,
				Edit:       true,
				AppendFunc: gormTagFunc(fieldNameMap, m, idm, opts),
			}, TagOption{
				Tag:   "sql",
				Type:  TypeSnakeCase,
				Cover: true,
				Edit:  true,
				AppendFunc: func(structName, fieldName, newTag, oldTag string) string {
					return strings.Join([]string{oldTag, sqlTagFunc(m)(structName, fieldName, newTag, oldTag)}, ";")
				},
			}
	}
	return TagOption{
			Tag:        opts.SqlTag,
			Type:       TypeSnakeCase,
			Cover:      true,
			Edit:       true,
			AppendFunc: gormTagFunc(fieldNameMap, m, idm, opts),
		}, TagOption{
			Tag:        "sql",
			Type:       TypeSnakeCase,
			Cover:      true,
			Edit:       true,
			AppendFunc: sqlTagFunc(m),
		}
}

// initConnect    初始化数据库连接.
func initConnect(dbConfig *DBConfig) *gorm.DB {
	dbConfig.Cstr = fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB,
	)
	var err error
	DB, err := gorm.Open(mysql.Open(dbConfig.Cstr), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		panic(err)
	}
	DB = DB.Debug()
	return DB
}

// index struct    索引定义.
type index struct {
	ColumnName string `gorm:"column:Column_name"` // 列名
	KeyName    string `gorm:"column:Key_name"`    // 索引名称
	NonUnique  int    `gorm:"column:Non_unique"`  // 是否唯一
}

// columnDef struct    表字段定义.
type columnDef struct {
	ColumnName    string `gorm:"column:COLUMN_NAME"`    // 字段名
	ColumnType    string `gorm:"column:COLUMN_TYPE"`    // 字段类型
	ColumnKey     string `gorm:"column:COLUMN_KEY"`     // 是否主键
	Extra         string `gorm:"column:EXTRA"`          // 额外的信息
	ColumnComment string `gorm:"column:COLUMN_COMMENT"` // 字段注释
	ColumnDefault string `gorm:"column:COLUMN_DEFAULT"` // 默认值
	IsNullable    string `gorm:"column:IS_NULLABLE"`    // 是否可空
}

// initTableDesc    初始化表结构.
func initTableDesc(table string, fieldNameMap map[string]string, db *gorm.DB) (m map[string]columnDef,
	idm map[string][]index) {
	var columns []columnDef
	err := db.Table("information_schema.COLUMNS").
		Where("table_name =  ?", table).Select("*").Find(&columns).Error
	if err != nil {
		panic(err)
	}
	rFieldNameMap := make(map[string]string)
	for k, v := range fieldNameMap {
		rFieldNameMap[v] = k
	}

	m = make(map[string]columnDef)
	for _, column := range columns {
		key := rFieldNameMap[column.ColumnName]
		if len(key) == 0 {
			key = column.ColumnName
		}
		m[key] = column
	}
	var indexes []index
	err = db.Raw("show index from `" + table + "`").Find(&indexes).Error
	if err != nil {
		if err.Error()[:10] == "Error 1146" {
			return m, idm
		}
		panic(err)
	}
	idm = make(map[string][]index)
	for _, index := range indexes {
		key := rFieldNameMap[index.ColumnName]
		if len(key) == 0 {
			key = index.ColumnName
		}
		idm[key] = append(idm[key], index)
	}
	return m, idm
}

// processPrimaryKey    处理主键属性.
func processPrimaryKey(t columnDef) []string {
	if t.ColumnKey == "PRI" {
		return []string{strings.ToUpper("primary_key")}
	}
	return []string{}
}

// processAutoIncrement    处理自增属性.
func processAutoIncrement(t columnDef) []string {
	if strings.Contains(t.Extra, "auto_increment") {
		return []string{strings.ToUpper("auto_increment")}
	}
	return []string{}
}

// processIndexes    处理索引属性.
func processIndexes(indexes []index) ([]string, []string) {
	var (
		indexList    []string
		unqIndexList []string
	)
	for _, i := range indexes {
		if i.KeyName != "PRIMARY" {
			if i.NonUnique == 1 {
				indexList = append(indexList, i.KeyName)
			} else {
				unqIndexList = append(unqIndexList, i.KeyName)
			}
		}
	}
	return indexList, unqIndexList
}

// buildGormProps    构建gorm属性.
func buildGormProps(opts *config.DbOpt, fieldNameMap map[string]string, fieldName string,
	t columnDef, indexes []index) []string {
	var props []string

	if opts.GormAnnotation {
		props = append(props, "column:"+fieldNameMap[fieldName])
	}

	props = append(props, processPrimaryKey(t)...)
	props = append(props, processAutoIncrement(t)...)
	props = append(props, "TYPE:"+t.ColumnType)

	if t.IsNullable == "NO" {
		props = append(props, "NOT NULL")
	}

	if len(indexes) > 0 {
		indexList, unqIndexList := processIndexes(indexes)
		if len(indexList) > 0 {
			props = append(props, "INDEX:"+strings.Join(indexList, ","))
		}
		if len(unqIndexList) > 0 {
			props = append(props, "UNIQUE_INDEX:"+strings.Join(unqIndexList, ","))
		}
	}

	return props
}

// gormTagFunc    生成gorm标签更新函数.
func gormTagFunc(fieldNameMap map[string]string, m map[string]columnDef,
	idm map[string][]index, opts *config.DbOpt) TagFunc {
	return func(structName, fieldName, newTag, oldTag string) (tag string) {
		t := m[fieldName]
		indexes := idm[fieldName]

		props := buildGormProps(opts, fieldNameMap, fieldName, t, indexes)
		return strings.Join(props, ";")
	}
}

// sqlTagFunc    生成sql标签更新函数.
func sqlTagFunc(m map[string]columnDef) func(structName, fieldName, newTag, oldTag string) (tag string) {
	return func(structName, fieldName, newTag, oldTag string) (tag string) {
		var props []string
		column := m[fieldName]
		if len(column.ColumnComment) > 0 {
			props = append(props, "COMMENT:'"+column.ColumnComment+"'")
		}
		if column.ColumnDefault != "" {
			d := "DEFAULT:" + column.ColumnDefault
			if column.Extra != "" {
				d += " " + column.Extra
			}
			props = append(props, d)
		}
		return strings.Join(props, ";")
	}
}
