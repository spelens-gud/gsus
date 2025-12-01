package generator

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spelens-gud/gsus/internal/config"
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

type Table struct {
	Name    string
	Comment string
}
type DBConfig struct {
	User     string
	Password string
	DB       string
	Host     string
	Port     int
	cstr     string
}
type columnDef struct {
	ColumnName    string `gorm:"column:COLUMN_NAME"`
	ColumnType    string `gorm:"column:COLUMN_TYPE"`
	ColumnKey     string `gorm:"column:COLUMN_KEY"`
	Extra         string `gorm:"column:EXTRA"`
	ColumnComment string `gorm:"column:COLUMN_COMMENT"`
	ColumnDefault string `gorm:"column:COLUMN_DEFAULT"`
	IsNullable    string `gorm:"column:IS_NULLABLE"`
}

type index struct {
	ColumnName string `gorm:"column:Column_name"`
	KeyName    string `gorm:"column:Key_name"`
	NonUnique  int    `gorm:"column:Non_unique"`
}
type TagFunc func(structName, fieldName, newTag, oldTag string) (tag string)

func GenAllDb2Struct(dir string, dbConfig DBConfig, options ...config.DbOption) (err error) {
	db, err := parser.GetConnection(dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB)
	if err != nil {
		return
	}

	tables, err := getTables(db, dbConfig.DB, "")
	if err != nil {
		return
	}

	for _, table := range tables {
		var ret []byte
		ret, err = GenTable(table.Name, dbConfig, options...)
		if err != nil {
			return err
		}

		if err = utils.ImportAndWrite(ret, filepath.Join(dir, strcase.SnakeCase(table.Name)+".go")); err != nil {
			return err
		}
	}
	return
}
func getTables(db *sql.DB, dbName string, tableName string) (ts []Table, err error) {
	sqlCommand := `SELECT TABLE_NAME, TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ?`
	args := []interface{}{dbName}
	if len(tableName) > 0 {
		sqlCommand += " AND TABLE_NAME = ?"
		args = append(args, tableName)
	}
	r, err := db.Query(sqlCommand, args...)
	if err != nil {
		return
	}
	// nolint
	defer r.Close()
	for r.Next() {
		var table Table
		if err = r.Scan(&table.Name, &table.Comment); err != nil {
			return
		}
		ts = append(ts, table)
	}
	return
}

func GenTable(table string, dbConfig DBConfig, options ...config.DbOption) (data []byte, err error) {
	// 原db2struct返回的顺序是字段字典序的 已改造成 db2structx
	db, err := parser.GetConnection(dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB)
	if err != nil {
		return
	}
	// nolint
	defer db.Close()

	res, err := parser.GetColumnsFromMysqlTable(db, dbConfig.DB, table)
	if err != nil {
		return
	}

	tables, err := getTables(db, dbConfig.DB, table)
	if err != nil || len(tables) != 1 {
		return
	}

	var (
		tableName  = tables[0]
		opts       = config.NewDbOpt(options...)
		structName = strcase.SnakeCase(table)
	)

	// 基础结构生成
	data, fieldNameMap := GenerateStruct(*res, table, structName, opts.PkgName, tableName.Comment, opts)

	// 结构体tag
	if data, err = parseTag(table, fieldNameMap, dbConfig, data, opts); err != nil {
		return
	}

	// 泛型方法
	genericF, err := parser.NewType(structName, opts.PkgName, opts.GenericOption...)
	if err != nil {
		return
	}

	data = append(data, genericF...)
	return
}

func GenerateStruct(columnTypes []parser.CTypes, tableName, structName, pkgName, tableComment string, opts *config.DbOpt) ([]byte, map[string]string) {
	dbTypes, fieldNameMap := generateMysqlTypes(columnTypes, 0, opts)
	src := fmt.Sprintf(template.HeadTemplate, tableComment, tableName, pkgName, tableName, ": "+tableComment, structName, dbTypes)
	tableNameFunc := fmt.Sprintf(template.TableFuncTemplate, structName, tableName, utils.GetFuncCallerIdent(structName), structName, structName)
	return []byte(src + "\n\n" + tableNameFunc), fieldNameMap
}

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

		// Get the corresponding go value type for this mysql type
		var valueType string
		// If the guregu (https://github.com/guregu/null) CLI option is passed use its types, otherwise use go's sql.NullX

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

func mysqlTypeToGoType(mysqlType string, nullable bool) string {
	switch mysqlType {
	case "tinyint", "int", "smallint", "mediumint":
		if nullable {
			return sqlNullInt
		}
		return golangInt
	case "json":
		return golangInterface
	case "bigint":
		if nullable {
			return sqlNullInt
		}
		return golangInt64
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
		if nullable {
			return sqlNullString
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		if nullable {
			return sqlNullTime
		}
		return golangTime
	case "decimal", "double":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat64
	case "float":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat32
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray
	}
	return golangInterface
}

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
func lintFieldName(name string) string {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}

	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}

	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		runes := []rune(name)
		if u := strings.ToUpper(name); commonInitialisms[u] {
			copy(runes[0:], []rune(u))
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		return string(runes)
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word

		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))

		} else if strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}

func stringifyFirstChar(str string) string {
	first := str[:1]

	i, err := strconv.ParseInt(first, 10, 8)

	if err != nil {
		return str
	}

	return intToWordMap[i] + "_" + str[1:]
}

func parseTag(table string, fieldNameMap map[string]string, dbConfig DBConfig, in []byte, opts *config.DbOpt) (data []byte, err error) {
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
		return
	}
	return
}
func GenTagUpdateFuncByTable(table string, fieldNameMap map[string]string, dbConfig DBConfig, opts *config.DbOpt) (gorm TagOption, sql TagOption) {
	db := initConnect(&dbConfig)
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

func initConnect(dbConfig *DBConfig) *gorm.DB {
	dbConfig.cstr = fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB,
	)
	var err error
	DB, err := gorm.Open(mysql.Open(dbConfig.cstr), &gorm.Config{
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

// TODO:使用原生SQL.
//
//nolint:godox
func initTableDesc(table string, fieldNameMap map[string]string, db *gorm.DB) (m map[string]columnDef, idm map[string][]index) {
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
			return
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
	return
}

func gormTagFunc(fieldNameMap map[string]string, m map[string]columnDef, idm map[string][]index, opts *config.DbOpt) TagFunc {
	return func(structName, fieldName, newTag, oldTag string) (tag string) {
		var props []string
		if opts.GormAnnotation {
			props = append(props, "column:"+fieldNameMap[fieldName])
		}
		t := m[fieldName]
		if t.ColumnKey == "PRI" {
			props = append(props, strings.ToUpper("primary_key"))
		}
		if strings.Contains(t.Extra, "auto_increment") {
			props = append(props, strings.ToUpper("auto_increment"))
		}
		props = append(props, "TYPE:"+t.ColumnType)
		if t.IsNullable == "NO" {
			props = append(props, "NOT NULL")
		}

		if len(idm[fieldName]) > 0 {
			var (
				indexes    []string
				unqIndexes []string
			)
			for _, i := range idm[fieldName] {
				if i.KeyName != "PRIMARY" {
					if i.NonUnique == 1 {
						indexes = append(indexes, i.KeyName)
					} else {
						unqIndexes = append(unqIndexes, i.KeyName)
					}
				}
			}
			if len(indexes) > 0 {
				props = append(props, "INDEX:"+strings.Join(indexes, ","))
			}
			if len(unqIndexes) > 0 {
				props = append(props, "UNIQUE_INDEX:"+strings.Join(unqIndexes, ","))
			}
		}
		tag = strings.Join(props, ";")
		return
	}
}

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
