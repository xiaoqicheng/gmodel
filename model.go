package gmodel

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

// ModelOptions .
type ModelOptions struct {
	MysqlDsn       string `json:"-" mapstructure:"dsn"` // mysql conn address
	MysqlTable     string `json:"-" mapstructure:"table"`
	Charset        string `json:"-" mapstructure:"charset"` // charset
	Collation      string `json:"-" mapstructure:"collation"`
	TablePrefix    string `json:"-" mapstructure:"table_prefix"`
	ColumnPrefix   string `json:"-" mapstructure:"column_prefix"`
	Package        string `json:"-" mapstructure:"pkg"`
	JSONTag        bool   `json:"-" mapstructure:"json_tag"`
	GormType       bool   `json:"-" mapstructure:"gorm_type"`
	ForceTableName bool   `json:"-" mapstructure:"with_table"`
	OutputPath     string `json:"-" mapstructure:"output_path"`
	SQL            string `json:"-"`
	InputFile      string `json:"-"`
	NoNullType     bool   `json:"-"`
	NullStyle      string `json:"-"`
	Update         bool   `json:"-"`
	Force          bool   `json:"-"`
}

var modelArgs = ModelOptions{}
var confOption = &ModelOptions{}

type GModelsConf struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type"`
}

const (
	defaultPath = "./"
	defaultName = "gmodel_config"
	defaultType = "yaml"
)

type options struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Option overrides behavior of conf.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

func WithGModelPath(path string) Option {
	return optionFunc(func(o *options) {
		o.Path = path
	})
}

func WithGModelName(name string) Option {
	return optionFunc(func(o *options) {
		o.Name = name
	})
}

func WithGModelType(t string) Option {
	return optionFunc(func(o *options) {
		o.Type = t
	})
}

// InitGModelConf Initialize the configuration file.
func InitGModelConf(opts ...Option) (*GModelsConf, error) {
	options := options{
		Path: defaultPath,
		Name: defaultName,
		Type: defaultType,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	return &GModelsConf{
		Path: options.Path,
		Name: options.Name,
		Type: options.Type,
	}, nil
}

// NewGModelCmd 获取一个 gorm model 生成 cmd
func (conf *GModelsConf) NewGModelCmd() *cobra.Command {
	conf.parseConfig()
	var modelCmd = &cobra.Command{
		Use:          "model",
		Short:        "generate model",
		Example:      "根据数据表自动生成model",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			modelTip()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			generateModel()
			return nil
		},
	}

	initParamsFlags(modelCmd)

	return modelCmd
}

func (conf *GModelsConf) parseConfig() {
	gmviper := viper.New()
	gmviper.AddConfigPath(conf.Path)
	gmviper.SetConfigName(conf.Name)
	gmviper.SetConfigType(conf.Type)

	if err := gmviper.ReadInConfig(); err != nil {
		log.Fatalf("Read config file error: %s", err)
		return
	}

	if err := gmviper.UnmarshalKey("gmodel", confOption); err != nil {
		log.Printf("Parse config.gmodel segment error: %s\n", err)
		return
	}
}
