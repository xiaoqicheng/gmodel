package gmodel

import (
	"github.com/spf13/cobra"
	"os"
	"testing"
)

func TestGeneralModel(t *testing.T) {
	if err := newGModel().Execute(); err != nil {
		os.Exit(1)
	}
}

func newGModel() *cobra.Command {
	gconf, _ := InitGModelConf(WithGModelPath("./"))
	return gconf.NewGModelCmd()
}
