package mount

import (
	"github.com/spelens-gud/gsus/apis/helpers/executor"
	"github.com/spelens-gud/gsus/apis/mount"
	"github.com/spelens-gud/gsus/internal/fileconfig"
	"github.com/spf13/cobra"
)

func Run(_ *cobra.Command, args []string) {
	executor.ExecuteWithConfig(func(_ fileconfig.Config) (err error) {
		argsMap := make(map[string]bool, len(args))
		for _, arg := range args {
			argsMap[arg] = true
		}
		return mount.Exec(mount.Config{
			Args: args,
		})
	})
}
