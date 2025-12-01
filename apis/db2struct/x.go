package db2struct

import (
	"fmt"
	"strings"

	"github.com/spelens-gud/gsus/apis/gengeneric"
	"github.com/spelens-gud/gsus/apis/tagger"
	"github.com/stoewer/go-strcase"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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

type DBConfig struct {
	User     string
	Password string
	DB       string
	Host     string
	Port     int

	cstr string
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

// TODO:使用原生SQL
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

func GenTagUpdateFuncByTable(table string, fieldNameMap map[string]string, dbConfig DBConfig, opts *opt) (gorm tagger.TagOption, sql tagger.TagOption) {
	db := initConnect(&dbConfig)
	m, idm := initTableDesc(table, fieldNameMap, db)
	if opts.commentOutside {
		for k, v := range m {
			v.ColumnComment = ""
			m[k] = v
		}
	}
	if opts.sqlTag == "sql" {
		return tagger.TagOption{
				Tag:        opts.sqlTag,
				Type:       tagger.TypeSnakeCase,
				Cover:      true,
				Edit:       true,
				AppendFunc: gormTagFunc(fieldNameMap, m, idm, opts),
			}, tagger.TagOption{
				Tag:   "sql",
				Type:  tagger.TypeSnakeCase,
				Cover: true,
				Edit:  true,
				AppendFunc: func(structName, fieldName, newTag, oldTag string) string {
					return strings.Join([]string{oldTag, sqlTagFunc(m)(structName, fieldName, newTag, oldTag)}, ";")
				},
			}
	}
	return tagger.TagOption{
			Tag:        opts.sqlTag,
			Type:       tagger.TypeSnakeCase,
			Cover:      true,
			Edit:       true,
			AppendFunc: gormTagFunc(fieldNameMap, m, idm, opts),
		}, tagger.TagOption{
			Tag:        "sql",
			Type:       tagger.TypeSnakeCase,
			Cover:      true,
			Edit:       true,
			AppendFunc: sqlTagFunc(m),
		}
}

func parseTag(table string, fieldNameMap map[string]string, dbConfig DBConfig, in []byte, opts *opt) (data []byte, err error) {
	gto, sto := GenTagUpdateFuncByTable(table, fieldNameMap, dbConfig, opts)
	pos := []tagger.TagOption{gto}
	if opts.sqlInfo {
		pos = append(pos, sto)
	}
	switch opts.json {
	case "camel":
		{
			pos = append(pos, tagger.CamelCase("json", false))
		}
	case "snake":
		{
			pos = append(pos, tagger.SnakeCase("json", false))
		}
	}
	data, _, err = tagger.ParseInput(in, pos...)
	if err != nil {
		return
	}
	return
}

func Gen(table string, dbConfig DBConfig, options ...Option) (data []byte, err error) {
	// 原db2struct返回的顺序是字段字典序的 已改造成 db2structx
	db, err := GetConnection(dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DB)
	if err != nil {
		return
	}
	// nolint
	defer db.Close()

	res, err := GetColumnsFromMysqlTable(db, dbConfig.DB, table)
	if err != nil {
		return
	}

	tables, err := getTables(db, dbConfig.DB, table)
	if err != nil || len(tables) != 1 {
		return
	}

	var (
		tableName  = tables[0]
		opts       = newOpt(options...)
		structName = strcase.SnakeCase(table)
	)

	// 基础结构生成
	data, fieldNameMap := Generate(*res, table, structName, opts.pkgName, tableName.Comment, opts)

	// 结构体tag
	if data, err = parseTag(table, fieldNameMap, dbConfig, data, opts); err != nil {
		return
	}

	// 泛型方法
	genericF, err := gengeneric.NewType(structName, opts.pkgName, opts.GenericOption...)
	if err != nil {
		return
	}

	data = append(data, genericF...)
	return
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

type TagFunc func(structName, fieldName, newTag, oldTag string) (tag string)

func gormTagFunc(fieldNameMap map[string]string, m map[string]columnDef, idm map[string][]index, opts *opt) TagFunc {
	return func(structName, fieldName, newTag, oldTag string) (tag string) {
		var props []string
		if opts.gormAnnotation {
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
