> ### The mysql model is command automatically generated and used with the gorm package
> If you like this gadget or have a good idea make a friend!
> **Thank you!**

> ## Quick Start
## Install

```shell
go get github.com/xiaoqicheng/gmodel
```

## What is gmodel?
gmodel is a command line service used to generate mysql data models, It's very rudimentary at the moment

It supports:
* reading from JSON, TOML, YAML, HCL, envfile config files as default value
* reading from command line flags

### The config.yaml example is as follows:

```yaml
gmodel:
  default:
    dsn: username:password@tcp(host:port)/database?charset=utf8&parseTime=True&loc=Asia%2FShanghai
    table: '*'
    pkg: internal   #要生成model所属包名
    with_table: true
    output_path: './dao/internal'  #输出model文件目录
    table_prefix: tbl_
    json_tag: true
    gorm_type: true
    unsigned: false #若为TRUE 则生成 uint类型; FALSE 为 int; default false
```

### add gmodel command

```go
var path = "./"
var name = "gmodel_config"
var ctype = "yaml"
var defaultMysql = "default"

func main() {
    rootCmd := &cobra.Command{}
    rootCmd.AddCommand(newGModel())
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func newGModel() *cobra.Command {
    gconf, _ := InitGModelConf(WithGModelConfPath(path), WithGModelConfName(name), WithGModelConfType(ctype), WithGModelConfDefaultMysql(defaultMysql))
    return gconf.NewGModelCmd()
}
```

You can use follow commands

```command
   create all new table model command; use "--slm" specified connection, if not used, default
   > go run main.go gmodel --slm default
    
   create a table model command 
   > go run main.go gmodel -t tablename
   
   update all table model command 
   > go run main.go gmodel -u -e
   
   update a table model command
   > go run main.go gmodel -u -t tablename
```