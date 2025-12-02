package generator

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spelens-gud/gsus/internal/parser"
	template2 "github.com/spelens-gud/gsus/internal/template"
	"golang.org/x/sync/errgroup"
)

var defaultRouterTemplate = template.Must(template.New("svc").Parse(template2.DefaultHttpRouterTemplate))

// GenApiRouterGroups function    生成接口路由组代码.
func GenApiRouterGroups(apiGroups []parser.ApiGroup, baseDir string, opts ...func(options *parser.GenOptions)) (err error) {
	if len(apiGroups) == 0 {
		return nil
	}

	o := &parser.GenOptions{
		Template:     defaultRouterTemplate,
		TemplateHash: fmt.Sprintf("%x", md5.Sum([]byte(template2.DefaultHttpRouterTemplate))),
	}

	for _, opt := range opts {
		opt(o)
	}
	wg := new(errgroup.Group)

	// 所有接口定义
	for i := range apiGroups {
		wg.Go(func() error {
			_ = os.MkdirAll(filepath.Dir(filepath.Join(baseDir, apiGroups[i].Filepath)), 0775)
			return o.WriteApiFiles(baseDir, &apiGroups[i])
		})
	}
	return wg.Wait()
}
