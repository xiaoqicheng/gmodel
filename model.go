package gmodel

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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
	Enforcement    bool   `json:"-"`
	JudgeUnsigned  bool   `json:"-" mapstructure:"unsigned"` //是否判断无符号 若为TRUE 则生成 uint类型; FALSE 为 int; default false
	SelectMySQL    string `json:"-"`                         //是否指定数据库
}

var modelArgs = ModelOptions{}
var confOption = &map[string]ModelOptions{}

type GModelsConf struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultMysql string `json:"default_mysql"`
}

const (
	defaultPath        = "./"
	defaultName        = "gmodel_config"
	defaultType        = "yaml"
	defaultSelectMysql = "default"
)

type options struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultMysql string `json:"default_mysql"`
}

// Option overrides behavior of conf.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

func WithGModelConfPath(path string) Option {
	return optionFunc(func(o *options) {
		if path != "" {
			o.Path = path
		}
		o.Path, _ = os.Getwd()
	})
}

func WithGModelConfName(name string) Option {
	return optionFunc(func(o *options) {
		o.Name = name
	})
}

func WithGModelConfType(t string) Option {
	return optionFunc(func(o *options) {
		o.Type = t
	})
}

func WithGModelConfDefaultMysql(selectMysql string) Option {
	return optionFunc(func(o *options) {
		o.DefaultMysql = selectMysql
	})
}

// InitGModelConf Initialize the configuration file.
func InitGModelConf(opts ...Option) (*GModelsConf, error) {
	options := options{
		Path:         defaultPath,
		Name:         defaultName,
		Type:         defaultType,
		DefaultMysql: defaultSelectMysql,
	}

	for _, o := range opts {
		o.apply(&options)
	}

	return &GModelsConf{
		Path:         options.Path,
		Name:         options.Name,
		Type:         options.Type,
		DefaultMysql: options.DefaultMysql,
	}, nil
}

// NewGModelCmd 获取一个 gorm model 生成 cmd
func (conf *GModelsConf) NewGModelCmd() *cobra.Command {
	if err := conf.parseConfig(); err != nil {
		fmt.Println(err)
		return &cobra.Command{}
	}

	var modelCmd = &cobra.Command{
		Use:          "gmodel",
		Short:        "generate model",
		Example:      "gmodel -t",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			modelTip()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			generateModel(conf.DefaultMysql)
			return nil
		},
	}

	conf.initParamsFlags(modelCmd)

	return modelCmd
}

func (conf *GModelsConf) parseConfig() error {
	gmviper := viper.New()
	gmviper.AddConfigPath(conf.Path)
	gmviper.SetConfigName(conf.Name)
	gmviper.SetConfigType(conf.Type)

	if err := gmviper.ReadInConfig(); err != nil {
		return fmt.Errorf("Read  gmodel cmd config file error: %s, so you can not use gmodel cmd", err)
	}

	if err := gmviper.UnmarshalKey("gmodel", confOption); err != nil {
		return fmt.Errorf("Parse config.gmodel segment error: %s\n", err)
	}

	if _, ok := (*confOption)[conf.DefaultMysql]; !ok {
		return fmt.Errorf("Parse config.gmodel.default  not exits")
	}

	return nil
}
