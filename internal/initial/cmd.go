package initial

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spelens-gud/gsus/apis/helpers"
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/basetmpl"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var Cmd = &cobra.Command{
	Use: "init",
	Run: run,
}

func run(_ *cobra.Command, args []string) {
	executor.Execute(func() (err error) {
		// try load config
		if _, err = fileconfig.Get(); err == nil {
			// config already exist
			err = fmt.Errorf("gsus has already init in project,run [ gsus update ] to update templates")
			return
		}

		// get project direction
		dir, err := helpers.GetProjectDir()
		if err != nil {
			return
		}

		// TODO: load config from origin
		// create .gsus dir
		_ = os.Mkdir(filepath.Join(dir, constant.ConfigDir), 0775)

		var tmp fileconfig.Config
		if err = yaml.Unmarshal([]byte(basetmpl.DefaultConfigYaml), &tmp); err != nil {
			return
		}
		bytes, err := yaml.Marshal(&tmp)
		if err != nil {
			return
		}

		// write default config
		if err = os.WriteFile(filepath.Join(dir, constant.ConfigFile), bytes, 0664); err != nil {
			return
		}

		// create templates dir
		_ = os.Mkdir(filepath.Join(dir, constant.TemplateDir), 0775)
		return writeTemplates(dir, templateMap)
	})
}

func writeTemplates(dir string, contentMap map[string]string) (err error) {
	for path, content := range contentMap {
		p := filepath.Join(dir, constant.TemplateDir, path+constant.TemplateSuffix)
		if err = os.WriteFile(p, []byte(content), 0664); err != nil {
			return
		}
	}
	return
}
