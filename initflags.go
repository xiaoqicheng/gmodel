package gmodel

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

// initParamsFlags .
func (conf *GModelsConf) initParamsFlags(modelCmd *cobra.Command) {

	//判断是否选定连接 -- 如果选定则使用，若没有选定则使用第一个连接
	modelCmd.Flags().StringVar(&modelArgs.SelectMySQL, "slm", conf.DefaultMysql, "table name prefix")

	defaultMysqlConf := (*confOption)[conf.DefaultMysql]

	modelCmd.Flags().StringVarP(&modelArgs.InputFile, "file", "f", "", "input file")
	modelCmd.Flags().StringVarP(&modelArgs.OutputPath, "output", "o", defaultMysqlConf.OutputPath, "output path")
	modelCmd.Flags().StringVarP(&modelArgs.SQL, "sql", "s", "", "input SQL")
	modelCmd.Flags().BoolVarP(&modelArgs.JSONTag, "json", "j", defaultMysqlConf.JSONTag, "generate json tag")
	modelCmd.Flags().StringVar(&modelArgs.TablePrefix, "table-prefix", defaultMysqlConf.TablePrefix, "table name prefix")
	modelCmd.Flags().StringVar(&modelArgs.ColumnPrefix, "col-prefix", "", "column name prefix")
	modelCmd.Flags().BoolVar(&modelArgs.NoNullType, "no-null", false, "do not use Null type")
	modelCmd.Flags().StringVar(&modelArgs.NullStyle, "null-style", "",
		"null type: sql.NullXXX(use 'sql') or *xxx(use 'ptr')")
	modelCmd.Flags().StringVarP(&modelArgs.Package, "pkg", "p", defaultMysqlConf.Package, "package name, default: model")
	modelCmd.Flags().BoolVar(&modelArgs.GormType, "with-type", defaultMysqlConf.GormType, "write type in gorm tag")
	modelCmd.Flags().BoolVar(&modelArgs.ForceTableName, "with-tablename", true, "write TableName func force")
	modelCmd.Flags().StringVarP(&modelArgs.MysqlDsn, "db-dsn", "d", defaultMysqlConf.MysqlDsn, "mysql dsn([user]:[pass]@tcp(host)/[database][?charset=xxx&...])")
	modelCmd.Flags().StringVarP(&modelArgs.MysqlTable, "db-table", "t", defaultMysqlConf.MysqlTable, "mysql table name")
	modelCmd.Flags().BoolVarP(&modelArgs.Update, "update", "u", false, "update table struct switch -t/-e")
	modelCmd.Flags().BoolVarP(&modelArgs.Enforcement, "enforcement", "e", false, "enforcement update all table struct switch -e")
	modelCmd.Flags().BoolVar(&modelArgs.JudgeUnsigned, "unsigned", false, "Whether to determine an unsigned type")
}

//如果不是默认的连接 则获取选定信息
func secondInitFlags(firstMysqlConf *ModelOptions, defaultSelectModel string) error {
	if _, ok := (*confOption)[modelArgs.SelectMySQL]; !ok {
		log.Printf("select mysql not exist")
		return fmt.Errorf("select mysql not exist")
	}

	selectMysqlConf := (*confOption)[modelArgs.SelectMySQL]
	defaultMysqlConf := (*confOption)[defaultSelectModel]

	if firstMysqlConf.InputFile == defaultMysqlConf.InputFile {
		firstMysqlConf.InputFile = selectMysqlConf.InputFile
	}
	if firstMysqlConf.OutputPath == defaultMysqlConf.OutputPath {
		firstMysqlConf.OutputPath = selectMysqlConf.OutputPath
	}
	if firstMysqlConf.SQL == defaultMysqlConf.SQL {
		firstMysqlConf.SQL = selectMysqlConf.SQL
	}
	if firstMysqlConf.JSONTag == defaultMysqlConf.JSONTag {
		firstMysqlConf.JSONTag = selectMysqlConf.JSONTag
	}
	if firstMysqlConf.TablePrefix == defaultMysqlConf.TablePrefix {
		firstMysqlConf.TablePrefix = selectMysqlConf.TablePrefix
	}
	if firstMysqlConf.ColumnPrefix == defaultMysqlConf.ColumnPrefix {
		firstMysqlConf.ColumnPrefix = selectMysqlConf.ColumnPrefix
	}
	if firstMysqlConf.NoNullType == defaultMysqlConf.NoNullType {
		firstMysqlConf.NoNullType = selectMysqlConf.NoNullType
	}
	if firstMysqlConf.NullStyle == defaultMysqlConf.NullStyle {
		firstMysqlConf.NullStyle = selectMysqlConf.NullStyle
	}
	if firstMysqlConf.Package == defaultMysqlConf.Package {
		firstMysqlConf.Package = selectMysqlConf.Package
	}
	if firstMysqlConf.GormType == defaultMysqlConf.GormType {
		firstMysqlConf.GormType = selectMysqlConf.GormType
	}
	if firstMysqlConf.ForceTableName == defaultMysqlConf.ForceTableName {
		firstMysqlConf.ForceTableName = selectMysqlConf.ForceTableName
	}
	if firstMysqlConf.MysqlDsn == defaultMysqlConf.MysqlDsn {
		firstMysqlConf.MysqlDsn = selectMysqlConf.MysqlDsn
	}
	if firstMysqlConf.MysqlTable == defaultMysqlConf.MysqlTable {
		firstMysqlConf.MysqlTable = selectMysqlConf.MysqlTable
	}

	if firstMysqlConf.Update == defaultMysqlConf.Update {
		firstMysqlConf.Update = selectMysqlConf.Update
	}
	if firstMysqlConf.Enforcement == defaultMysqlConf.Enforcement {
		firstMysqlConf.Enforcement = selectMysqlConf.Enforcement
	}
	if firstMysqlConf.JudgeUnsigned == defaultMysqlConf.JudgeUnsigned {
		firstMysqlConf.JudgeUnsigned = selectMysqlConf.JudgeUnsigned
	}

	return nil
}
