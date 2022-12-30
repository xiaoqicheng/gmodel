package gmodel

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xiaoqicheng/gmodel/color"
	"github.com/xiaoqicheng/gmodel/parser"
	"log"
	"os"
	"strings"
	"sync"
)

// initParamsFlags .
func initParamsFlags(modelCmd *cobra.Command) {

	//判断是否选定连接 -- 如果选定则使用，若没有选定则使用第一个连接
	modelCmd.Flags().StringVar(&modelArgs.SelectMySQL, "slm", defaultSelectMysql, "table name prefix")

	if _, ok := (*confOption)[modelArgs.SelectMySQL]; !ok {
		log.Printf("select mysql not exist")
		return
	}

	selectMysqlConf := (*confOption)[modelArgs.SelectMySQL]

	modelCmd.Flags().StringVarP(&modelArgs.InputFile, "file", "f", "", "input file")
	modelCmd.Flags().StringVarP(&modelArgs.OutputPath, "output", "o", selectMysqlConf.OutputPath, "output path")
	modelCmd.Flags().StringVarP(&modelArgs.SQL, "sql", "s", "", "input SQL")
	modelCmd.Flags().BoolVarP(&modelArgs.JSONTag, "json", "j", selectMysqlConf.JSONTag, "generate json tag")
	modelCmd.Flags().StringVar(&modelArgs.TablePrefix, "table-prefix", selectMysqlConf.TablePrefix, "table name prefix")
	modelCmd.Flags().StringVar(&modelArgs.ColumnPrefix, "col-prefix", "", "column name prefix")
	modelCmd.Flags().BoolVar(&modelArgs.NoNullType, "no-null", false, "do not use Null type")
	modelCmd.Flags().StringVar(&modelArgs.NullStyle, "null-style", "",
		"null type: sql.NullXXX(use 'sql') or *xxx(use 'ptr')")
	modelCmd.Flags().StringVarP(&modelArgs.Package, "pkg", "p", selectMysqlConf.Package, "package name, default: model")
	modelCmd.Flags().BoolVar(&modelArgs.GormType, "with-type", selectMysqlConf.GormType, "write type in gorm tag")
	modelCmd.Flags().BoolVar(&modelArgs.ForceTableName, "with-tablename", true, "write TableName func force")
	modelCmd.Flags().StringVarP(&modelArgs.MysqlDsn, "db-dsn", "d", selectMysqlConf.MysqlDsn, "mysql dsn([user]:[pass]@tcp(host)/[database][?charset=xxx&...])")
	modelCmd.Flags().StringVarP(&modelArgs.MysqlTable, "db-table", "t", selectMysqlConf.MysqlTable, "mysql table name")
	modelCmd.Flags().BoolVarP(&modelArgs.Update, "update", "u", false, "update table struct switch -t/-e")
	modelCmd.Flags().BoolVarP(&modelArgs.Enforcement, "enforcement", "e", false, "enforcement update all table struct switch -e")
	modelCmd.Flags().BoolVar(&modelArgs.JudgeUnsigned, "unsigned", false, "Whether to determine an unsigned type")
}

// getOptions .
func getOptions(args ModelOptions) []parser.Option {
	opt := make([]parser.Option, 0, 1)
	if args.Charset != "" {
		opt = append(opt, parser.WithCharset(args.Charset))
	}
	if args.Collation != "" {
		opt = append(opt, parser.WithCollation(args.Collation))
	}
	if args.JSONTag {
		opt = append(opt, parser.WithJSONTag())
	}
	if args.TablePrefix != "" {
		opt = append(opt, parser.WithTablePrefix(args.TablePrefix))
	}
	if args.ColumnPrefix != "" {
		opt = append(opt, parser.WithColumnPrefix(args.ColumnPrefix))
	}
	if args.NoNullType {
		opt = append(opt, parser.WithNoNullType())
	}
	if args.NullStyle != "" {
		switch args.NullStyle {
		case "sql":
			opt = append(opt, parser.WithNullStyle(parser.NullInSQL))
		case "ptr":
			opt = append(opt, parser.WithNullStyle(parser.NullInPointer))
		default:
			fmt.Printf("invalid null style: %s\n", args.NullStyle)
			return nil
		}
	}

	if args.Package != "" {
		opt = append(opt, parser.WithPackage(args.Package))
	}
	if args.GormType {
		opt = append(opt, parser.WithGormType())
	}
	if args.ForceTableName {
		opt = append(opt, parser.WithForceTableName())
	}

	if args.JudgeUnsigned {
		opt = append(opt, parser.WithJudgeUnsigned())
	}
	return opt
}

// GenerateModel .
func generateModel() {

	judgeUpdateArgs()
	//判断mysql连接是否为空参数
	judgeMysqlDsnIsNull()
	//sql 获取顺序为： -s > -f > "自动获取"
	judgeMysqlSqlWithTable()

	// 获取即将生成的表结构的所有表
	tables, err := parser.GetCreateTables(modelArgs.MysqlDsn, modelArgs.MysqlTable)
	if err != nil {
		exitWithInfo("get tables error: %s", err)
	}

	// 已存在的表不在更新， 只新增不存在的表， 除非使用 更新命令
	wg := &sync.WaitGroup{}
	for _, table := range tables {
		wg.Add(1)
		go writeModelFile(table, modelArgs.SQL, wg)
	}

	wg.Wait()
	fmt.Printf("%s \n", color.Blue(`success`))
	return
}

// createModelFile 创建model file 在指定目录
func createModelFile(tableName, tablePrefix, filePath string) (*os.File, bool) {
	fileName := tableName
	if modelArgs.TablePrefix != "" && strings.HasPrefix(fileName, tablePrefix) {
		fileName = fileName[len(tablePrefix):]
	}

	fileAddress := filePath + "/" + fileName + ".go"
	//判断 -update 参数 是否存在 是 则不判断; 否 则判断文件是否存在; -u -t 更新某个表model; -u -e 更新全部表model
	if modelArgs.Update == false {
		//如果文件存在 则 跳过
		if ok, _ := pathExists(fileAddress); ok {
			fmt.Println(color.Cyan("model已存在 [" + tableName + "]"))
			return nil, ok
		}
	}

	f, err := os.OpenFile(fileAddress, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		exitWithInfo("open %s failed, %s\n", filePath, err)
	}

	return f, false
}

//writeModelFile 将 model 写入 文件
func writeModelFile(table, sql string, wg *sync.WaitGroup) {
	defer wg.Done()
	//确定输出目录
	dirPath, err := initDirPath(modelArgs.OutputPath)
	if err != nil {
		exitWithInfo("init dir path %s failed, %s\n", modelArgs.OutputPath, err)
	}

	f, ok := createModelFile(table, modelArgs.TablePrefix, dirPath)
	defer f.Close()

	//如果文件已存在 则 跳过
	if ok {
		return
	}

	opt := getOptions(modelArgs)
	if opt == nil {
		return
	}

	fmt.Println(color.Yellow("正在生成 [" + table + "]"))
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(color.Red("生成错误 [" + table + "]" + fmt.Sprintf("%v", err)))
			return
		}

		fmt.Println(color.Green("生成完毕 [" + table + "]"))
	}()

	if sql == "" {
		//自动获取 sql 参数
		sql, err = parser.GetCreateTableFromDB(modelArgs.MysqlDsn, table)
		if err != nil {
			exitWithInfo("get create table error: %s", err)
		}
	}

	err = parser.ParseSQLToWrite(sql, f, opt...)
	if err != nil {
		exitWithInfo(err.Error())
	}
}

// initDirPath .
func initDirPath(dirPath string) (string, error) {
	if dirPath == "" {
		dirPath = "./"
	} else {
		if ok, _ := pathExists(dirPath); !ok {
			err := os.MkdirAll(dirPath, os.ModePerm)
			return "", err
		}
	}
	return dirPath, nil
}

// PathExists .
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// modelTip
func modelTip() {
	fmt.Printf("%s\n", `欢迎使用 `+color.Green(`根据数据表自动生成model`)+` 可以使用 `+color.Red(`-h`)+` 查看命令`)
	fmt.Printf("%s\n", color.Blue(`In the making...`))
}

// ExitWithInfo .
func exitWithInfo(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

//judgeUpdateArgs 判断 update 命令是否配合 -t / -e
func judgeUpdateArgs() {
	if modelArgs.Update {
		if !modelArgs.Enforcement && (modelArgs.MysqlTable == "" || modelArgs.MysqlTable == "*") {
			_, _ = fmt.Fprintf(os.Stderr, "no table or enforcement input(-t|-e)\n\n")
			flag.Usage()
			os.Exit(2)
		}
	}
}

//judgeMysqlDsnIsNull 判断 dsn 连接是否为空
func judgeMysqlDsnIsNull() {
	if modelArgs.MysqlDsn == "" {
		exitWithInfo("miss mysql conn, please add a configuration")
	}
}

//judgeMysqlSqlWithTable 获取sql 并 判断 sql 是否配合-t使用
func judgeMysqlSqlWithTable() {
	if modelArgs.SQL == "" {
		if modelArgs.InputFile != "" {
			b, err := os.ReadFile(modelArgs.InputFile)
			if err != nil {
				exitWithInfo("read %s failed, %s\n", modelArgs.InputFile, err)
			}
			modelArgs.SQL = string(b)
		}
	}

	//如果指定 sql 语句；则 -t 必须存在， -t 存在 sql 不一定存在;  不支持 一个 sql 多个table的用法
	if modelArgs.SQL != "" {
		if modelArgs.MysqlTable == "" || modelArgs.MysqlTable == "*" {
			exitWithInfo("you need use -t for a table name")
		}
		return
	}
}
