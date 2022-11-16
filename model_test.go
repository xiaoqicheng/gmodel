package gmodel

import (
	"bytes"
	"github.com/spf13/cobra"
	"testing"
)

var path = "./"
var name = "gmodel_config"
var ctype = "yaml"

func TestGeneralModel(t *testing.T) {
	rootCmd := &cobra.Command{}

	rootCmd.AddCommand(newGModel())

	output, err := executeCommand(rootCmd, "gmodel")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func newGModel() *cobra.Command {
	gconf, _ := InitGModelConf(WithGModelConfPath(path), WithGModelConfName(name), WithGModelConfType(ctype))
	return gconf.NewGModelCmd()
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
