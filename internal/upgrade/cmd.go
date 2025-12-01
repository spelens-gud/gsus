package upgrade

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spelens-gud/gsus/apis/constant"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "upgrade",
	Run:   run,
	Short: "update go-gsus to latest version",
}

var vMode *bool

func init() {
	vMode = Cmd.Flags().Bool("v", false, "log more detail")
}

func run(_ *cobra.Command, _ []string) {
	vStr := ""
	if *vMode {
		vStr = "-v"
	}

	log.Printf("updating gsus from [ %s ] ...", constant.GoGetUrl)
	if err := execCommand("go", "get", "-u", vStr, constant.GoGetUrl); err != nil {
		if strings.Contains(err.Error(), "uses insecure protocol") {
			if err = execCommand("go", "get", "--insecure", "-u", vStr, constant.GoGetUrl); err == nil {
				return
			}
		}
		log.Fatal(err)
	}
}

func execCommand(name string, commands ...string) (err error) {
	var j int
	for _, c := range commands {
		if len(c) > 0 {
			commands[j] = c
			j++
		}
	}
	commands = commands[:j]
	command := exec.Command(name, commands...)
	command.Stdout = os.Stdout
	bf := new(bytes.Buffer)
	command.Stderr = io.MultiWriter(os.Stderr, bf)
	if err = command.Run(); err != nil {
		err = fmt.Errorf("%s: %v", bf.String(), err)
		return
	}
	return
}
