package helpers

import (
	"sync"

	"github.com/Just-maple/xtoolinternal/gocommand"
	"github.com/Just-maple/xtoolinternal/imports"
	imports2 "golang.org/x/tools/imports"
)

var importMu sync.Mutex

func ImportProcess(bytes []byte) (ret []byte, err error) {
	importMu.Lock()
	defer importMu.Unlock()
	return imports.Process("", bytes, opt)
}

var localPrefix = func() string {
	path, _ := GetModBase()
	return path
}()

var (
	opt2 = &imports2.Options{Comments: true, TabIndent: true, TabWidth: 8}
	opt  = &imports.Options{
		Env: &imports.ProcessEnv{
			GocmdRunner: &gocommand.Runner{},
		},
		LocalPrefix: localPrefix,
		AllErrors:   opt2.AllErrors,
		Comments:    opt2.Comments,
		FormatOnly:  opt2.FormatOnly,
		Fragment:    opt2.Fragment,
		TabIndent:   opt2.TabIndent,
		TabWidth:    opt2.TabWidth,
	}
)
