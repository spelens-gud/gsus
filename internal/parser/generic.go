package parser

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spelens-gud/gsus/internal/config"
	"github.com/spelens-gud/gsus/internal/errors"
	template2 "github.com/spelens-gud/gsus/internal/template"
	"github.com/spelens-gud/gsus/internal/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var defaultTemplate *template.Template

type T struct {
	MapTypes []MapType
	Type     string
	Package  string
}

type MapType struct {
	MapType  string
	MapBType string
}

func init() {
	defaultTemplate = template.Must(template.New("").Parse(template2.DefaultModelGenericTemplate))
}

func NewType(typeName, pkg string, opts ...func(o *config.Options)) (res string, err error) {
	t := T{
		Type:    typeName,
		Package: pkg,
	}
	o := &config.Options{
		MapTypes: []string{"int", "string"},
		Template: defaultTemplate,
	}
	for _, opt := range opts {
		opt(o)
	}

	for _, typ := range o.MapTypes {
		caser := cases.Title(language.English)
		t.MapTypes = append(t.MapTypes, MapType{
			MapType:  caser.String(typ),
			MapBType: typ,
		})
	}

	ret, err := utils.ExecuteTemplate(o.Template, t)
	if err != nil {
		return res, errors.WrapWithCode(err, errors.ErrCodeTemplate, fmt.Sprintf("模板执行失败: %s", err))
	}
	res = string(ret)
	return
}

type GenOptions struct {
	Template      *template.Template
	TemplateHash  string
	DisableSkip   bool
	OverwriteTest bool
}

func (opt *GenOptions) WriteApiFiles(baseDir string, route *ApiGroup) (err error) {
	defer func() {
		if err != nil {
			log.Printf("generate http router error [ %s ]:%v", route.Filepath, err)
		}
	}()

	// 检查hash
	hashBytes, _ := json.Marshal(route)
	route.Hash = fmt.Sprintf("%x", md5.Sum(append(hashBytes, opt.TemplateHash...)))

	fp := filepath.Join(baseDir, route.Filepath)
	if b, hasErr := os.ReadFile(fp); hasErr == nil {
		if strings.Contains(string(b), route.Hash) {
			if !opt.DisableSkip {
				route.Skip = true
			}
			log.Printf("generate [ %s ] hash unchanged,skip", route.Filepath)
			return errors.New(errors.ErrCodeGenerate, "skip")
		}
	}

	log.Printf("generating http router [ %s ]", route.Filepath)
	if err = utils.ExecuteTemplateAndWrite(opt.Template, route, fp); err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeGenerate, fmt.Sprintf("生成文件 [ %s ] 失败", route.Filepath))
	}
	return
}
