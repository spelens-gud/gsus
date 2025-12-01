package routergen

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/httpgen"
	"github.com/spelens-gud/gsus/basetmpl"
	"golang.org/x/sync/errgroup"
)

type GenOptions struct {
	Template      *template.Template
	TemplateHash  string
	DisableSkip   bool
	OverwriteTest bool
}

func GenApiRouterGroups(apiGroups []httpgen.ApiGroup, baseDir string, opts ...func(options *GenOptions)) (err error) {
	if len(apiGroups) == 0 {
		return
	}

	o := &GenOptions{
		Template:     defaultRouterTemplate,
		TemplateHash: fmt.Sprintf("%x", md5.Sum([]byte(basetmpl.DefaultHttpRouterTemplate))),
	}

	for _, opt := range opts {
		opt(o)
	}
	wg := new(errgroup.Group)

	// 所有接口定义
	for i := range apiGroups {
		i := i
		wg.Go(func() error {
			_ = os.MkdirAll(filepath.Dir(filepath.Join(baseDir, apiGroups[i].Filepath)), 0775)
			return o.WriteApiFiles(baseDir, &apiGroups[i])
		})
	}
	return wg.Wait()
}

func (opt *GenOptions) WriteApiFiles(baseDir string, route *httpgen.ApiGroup) (err error) {
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
			return
		}
	}

	log.Printf("generating http router [ %s ]", route.Filepath)
	if err = helpers.ExecuteTemplateAndWrite(opt.Template, route, fp); err != nil {
		return
	}
	return
}
