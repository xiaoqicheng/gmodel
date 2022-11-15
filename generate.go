package gmodel

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xiaoqicheng/gmodel/color"
	"github.com/xiaoqicheng/gmodel/parser"
	"os"
	"strings"
	"sync"
	"unicode"
)

// initParamsFlags .
func initParamsFlags(modelCmd *cobra.Command) {
	modelCmd.Flags().StringVarP(&modelArgs.InputFile, "file", "f", "", "input file")
	modelCmd.Flags().StringVarP(&modelArgs.OutputPath, "output", "o", confOption.OutputPath, "output path")
	modelCmd.Flags().StringVarP(&modelArgs.SQL, "sql", "s", "", "input SQL")
	modelCmd.Flags().BoolVarP(&modelArgs.JSONTag, "json", "j", confOption.JSONTag, "generate json tag")
	modelCmd.Flags().StringVar(&modelArgs.TablePrefix, "table-prefix", confOption.TablePrefix, "table name prefix")
	modelCmd.Flags().StringVar(&modelArgs.ColumnPrefix, "col-prefix", "", "column name prefix")
	modelCmd.Flags().BoolVar(&modelArgs.NoNullType, "no-null", false, "do not use Null type")
	modelCmd.Flags().StringVar(&modelArgs.NullStyle, "null-style", "",
		"null type: sql.NullXXX(use 'sql') or *xxx(use 'ptr')")
	modelCmd.Flags().StringVarP(&modelArgs.Package, "pkg", "p", confOption.Package, "package name, default: model")
	modelCmd.Flags().BoolVar(&modelArgs.GormType, "with-type", confOption.GormType, "write type in gorm tag")
	modelCmd.Flags().BoolVar(&modelArgs.ForceTableName, "with-tablename", true, "write TableName func force")
	modelCmd.Flags().StringVarP(&modelArgs.MysqlDsn, "db-dsn", "d", confOption.MysqlDsn, "mysql dsn([user]:[pass]@tcp(host)/[database][?charset=xxx&...])")
	modelCmd.Flags().StringVarP(&modelArgs.MysqlTable, "db-table", "t", confOption.MysqlTable, "mysql table name")
	modelCmd.Flags().BoolVarP(&modelArgs.Update, "update", "u", false, "if true update all table struct")
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
	return opt
}

// GenerateModel .
func generateModel() {
	//判断mysql连接是否为空参数
	if modelArgs.MysqlDsn == "" {
		exitWithInfo("miss mysql conn, please add a configuration")
	}

	//sql 获取顺序为： -s > -f > "自动获取"
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

	// 获取即将生成的表结构的所有表
	tables, err := parser.GetCreateTables(modelArgs.MysqlDsn, modelArgs.MysqlTable)
	if err != nil {
		exitWithInfo("get tables error: %s", err)
	}

	// 已存在的表不在更新， 只新增不存在的表， 除非使用 更新命令
	wg := &sync.WaitGroup{}
	//ch := make(chan struct{}, runtime.NumCPU())
	for _, table := range tables {
		wg.Add(1)
		//ch <- struct{}{}
		go writeModelFile(table, modelArgs.SQL, wg)
	}

	wg.Wait()
	usageStr := color.Blue(`success`)
	fmt.Printf("%s \n", usageStr)
	return
}

// createModelFile 创建model file 在指定目录
func createModelFile(tableName, tablePrefix, filePath string) (*os.File, bool) {
	fileName := tableName
	if modelArgs.TablePrefix != "" && strings.HasPrefix(fileName, tablePrefix) {
		tableNameTemp := fileName[len(tablePrefix):]
		for _, v := range tableNameTemp {
			if !unicode.IsNumber(v) {
				fileName = tableNameTemp
			}
			break
		}
	}

	fileAddress := filePath + "/" + fileName + ".go"
	//判断 -update 参数 是否存在 是 则不判断; 否 则判断文件是否存在; -u  -t 更新某个表model
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
