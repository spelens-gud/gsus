package generator

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/spelens-gud/gsus/internal/errors"
	"github.com/stoewer/go-strcase"
)

func tLog(f string, arg ...interface{}) {
	fmt.Printf(f+"\n", arg...)
}

// ParseAllPath function    递归解析所有路径下的 Go 文件标签.
func ParseAllPath(file string, opts ...TagOption) (err error) {
	dir, err := os.ReadDir(file)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("读取目录失败: %s", err))
	}
	for _, f := range dir {
		if f.IsDir() {
			if f.Name() == "vendor" {
				continue
			}
			err = ParseAllPath(file+"/"+f.Name(), opts...)
			if err != nil {
				return errors.Wrap(err, "递归解析目录失败")
			}
			continue
		} else if f.Name()[len(f.Name())-3:] == ".go" {
			err = ParseTag(file+"/"+f.Name(), opts...)
			if err != nil {
				return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("解析标签失败: %s", err))
			}
		}
	}
	return
}

// ParsePath function    解析指定路径下的 Go 文件标签.
func ParsePath(file string, opts ...TagOption) (err error) {
	dir, err := os.ReadDir(file)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("读取目录失败; %s", err))
	}
	for _, f := range dir {
		if f.IsDir() {
			continue
		}
		if f.Name()[len(f.Name())-3:] == ".go" {
			err = ParseTag(file+"/"+f.Name(), opts...)
			if err != nil {
				return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("解析标签失败: %s", err))
			}
		}
	}
	return
}

type objs []obj
type obj struct {
	// lint:ignore SA1019 理由
	// nolint
	o *ast.Object
	p int
}

func (o objs) Len() int {
	return len(o)
}

func (o objs) Less(i, j int) bool {
	return o[i].p < o[j].p
}

func (o objs) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

// ParseTag function    解析单个文件的标签.
func ParseTag(file string, opts ...TagOption) (err error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("读取文件失败: %s", err))
	}
	res, edit, err := ParseInput(data, opts...)
	if err != nil || !edit {
		return errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("解析标签失败: %s", err))
	}
	err = os.WriteFile(file, res, os.FileMode(0664))
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("写入文件失败: %s", err))
	}
	return
}

// ParseInput function    解析输入数据的标签.
func ParseInput(data []byte, opts ...TagOption) (res []byte, edited bool, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", data, 0)
	if err != nil {
		return nil, false, errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("解析文件失败: %s", err))
	}
	var istp int
	t := len(data)
	var appendTag = func(f *ast.Field, content []byte) {
		tmp := make([]byte, len(data[int(f.Type.End())+istp-1:]))
		copy(tmp, data[int(f.Type.End())+istp-1:])
		str := content
		data = append(data[:int(f.Type.End())+istp-1], str...)
		data = append(data, tmp...)
		istp += len(str)
	}
	var replaceTag = func(f *ast.Field, content []byte) {
		old := data[f.Tag.Pos() : f.Tag.End()-2]
		prel := len(old)
		tmp := make([]byte, len(data[int(f.Tag.End())+istp-2:]))
		copy(tmp, data[int(f.Tag.End())+istp-2:])
		data = append(data[:int(f.Tag.Pos())+istp], content...)
		data = append(data, tmp...)
		istp += len(content) - prel
	}
	var objects objs
	for _, i := range f.Scope.Objects {
		objects = append(objects, obj{
			o: i,
			p: int(i.Pos()),
		})
	}
	sort.Sort(objects)
	for _, d := range objects {
		structObj := d.o
		ts, ok := structObj.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			continue
		}
		if st.Fields == nil {
			continue
		}
		for _, f := range st.Fields.List {
			var edited bool
			if len(f.Names) == 0 || (f.Names[0].Name[0] <= 'z' && f.Names[0].Name[0] >= 'a') {
				continue
			}
			fieldName := ""
			for _, n := range f.Names {
				fieldName += n.Name
			}
			if f.Tag != nil {
				newTag := genTag(f, fieldName, structObj.Name, opts...)
				s := string(data[int(f.Tag.Pos())+istp : int(f.Tag.End())+istp-2])
				if !strings.EqualFold(string(newTag), s) {
					edited = true
				}
				replaceTag(f, newTag)
				//	已有tag不管
			} else {
				// 匿名字段 不管
				edited = true
				appendTag(f, []byte(" "+fmt.Sprintf("`%s`", genTag(f, fieldName, structObj.Name, opts...))))
			}
			if edited {
				tLog("parse %s %s", structObj.Name, fieldName)
			}
		}
	}
	if istp == 0 {
		res = data
		return
	}
	edited = true
	data = data[:t+istp]
	res, err = format.Source(data)
	if err != nil {
		return nil, false, errors.WrapWithCode(err, errors.ErrCodeParse, fmt.Sprintf("格式化源代码失败: %s", err))
	}
	return
}

var tagRegexp = regexp.MustCompile(`(.+?):"(.*?)"`)

func genTag(f *ast.Field, name, structName string, opts ...TagOption) []byte {
	var tagMap = make(map[string]string)
	var tags []string
	if f.Tag != nil {
		oldTags := tagRegexp.FindAllStringSubmatch(f.Tag.Value[1:len(f.Tag.Value)-1], -1)
		for _, tag := range oldTags {
			if len(tag) != 3 {
				continue
			}
			tagMap[strings.TrimSpace(tag[1])] = strings.TrimSpace(tag[2])
			tags = append(tags, strings.TrimSpace(tag[1]))
		}
	}
	for _, opt := range opts {
		var addTag = func() {
			tag := tagMap[opt.Tag]
			switch opt.Type {
			case TypeCamelCase:
				tagMap[opt.Tag] = strcase.LowerCamelCase(name)
			case TypeSnakeCase:
				tagMap[opt.Tag] = strcase.SnakeCase(name)
			}
			if opt.AppendFunc != nil {
				res := opt.AppendFunc(structName, name, tagMap[opt.Tag], tag)
				if res != "" {
					tagMap[opt.Tag] = res
				} else {
					delete(tagMap, opt.Tag)
				}
			}
		}
		_, ok := tagMap[opt.Tag]
		if ok {
			if opt.Cover {
				addTag()
			}
			continue
		}
		if f.Tag != nil && !opt.Edit {
			continue
		}
		tags = append(tags, opt.Tag)
		addTag()
	}
	var res []string
	for _, tag := range tags {
		if tagMap[tag] != "" {
			res = append(res, tag+`:"`+tagMap[tag]+`"`)
		}
	}
	return []byte(strings.Join(res, " "))
}

const (
	TypeCamelCase = 1
	TypeSnakeCase = 0
)

type TagOption struct {
	Tag        string                                                    `json:"tag" gorm:"column:tag" bson:"tag" form:"tag"`
	Type       int                                                       `json:"type" gorm:"column:type" bson:"type" form:"type"`
	Cover      bool                                                      `json:"cover" gorm:"column:cover" bson:"cover" form:"cover"` // cover old tag
	Edit       bool                                                      `json:"edit" gorm:"column:edit" bson:"edit" form:"edit"`     // edit tag
	AppendFunc func(structName, fieldName, newTag, oldTag string) string `json:"append_func" form:"append_func" gorm:"column:append_func" bson:"append_func"`
}

func CamelCase(tag string, cover bool) TagOption {
	return TagOption{
		Type:  TypeCamelCase,
		Tag:   tag,
		Cover: cover,
		Edit:  true,
	}
}

func SnakeCase(tag string, cover bool) TagOption {
	return TagOption{
		Type:  TypeSnakeCase,
		Tag:   tag,
		Cover: cover,
		Edit:  true,
	}
}
