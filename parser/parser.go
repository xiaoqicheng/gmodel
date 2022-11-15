package parser

import (
	"fmt"
	"go/format"
	"io"
	"sort"
	"strings"
	"sync"
	"text/template"
	"unicode"

	"github.com/blastrain/vitess-sqlparser/tidbparser/ast"
	"github.com/blastrain/vitess-sqlparser/tidbparser/dependency/mysql"
	"github.com/blastrain/vitess-sqlparser/tidbparser/dependency/types"
	"github.com/blastrain/vitess-sqlparser/tidbparser/parser"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
)

var (
	structTmplRaw string
	fileTmplRaw   string
	structTmpl    *template.Template
	fileTmpl      *template.Template
	tmplParseOnce sync.Once
)

var acronym = map[string]struct{}{
	"ID":  {},
	"IP":  {},
	"RPC": {},
}

// ModelCodes .
type ModelCodes struct {
	Package    string   `json:"-"`
	ImportPath []string `json:"-"`
	StructCode []string `json:"-"`
}

// ParseSQL .
func ParseSQL(sql string, options ...Option) (ModelCodes, error) {
	initTemplate()
	opt := parseOption(options)

	stmts, err := parser.New().Parse(sql, opt.Charset, opt.Collation)
	if err != nil {
		return ModelCodes{}, err
	}
	tableStr := make([]string, 0, len(stmts))
	importPath := make(map[string]struct{})
	for _, stmt := range stmts {
		if ct, ok := stmt.(*ast.CreateTableStmt); ok {
			s, ipt, err := makeCode(ct, opt)
			if err != nil {
				return ModelCodes{}, err
			}
			tableStr = append(tableStr, s)
			for _, s := range ipt {
				importPath[s] = struct{}{}
			}
		}
	}

	importPathArr := make([]string, 0, len(importPath))
	for s := range importPath {
		importPathArr = append(importPathArr, s)
	}
	sort.Strings(importPathArr)

	return ModelCodes{
		Package:    opt.Package,
		ImportPath: importPathArr,
		StructCode: tableStr,
	}, nil
}

// ParseSQLToWrite .
func ParseSQLToWrite(sql string, writer io.Writer, options ...Option) error {
	data, err := ParseSQL(sql, options...)
	if err != nil {
		return err
	}
	err = fileTmpl.Execute(writer, data)
	if err != nil {
		return err
	}
	return nil
}

// ConfigureAcronym .
func ConfigureAcronym(words []string) {
	acronym = make(map[string]struct{}, len(words))
	for _, w := range words {
		acronym[strings.ToUpper(w)] = struct{}{}
	}
}

// tmplData .
type tmplData struct {
	TableName    string      `json:"-"`
	NameFunc     bool        `json:"-"`
	RawTableName string      `json:"-"`
	Fields       []tmplField `json:"-"`
	Comment      string      `json:"-"`
}

// tmplField .
type tmplField struct {
	Name    string `json:"-"`
	GoType  string `json:"-"`
	Tag     string `json:"-"`
	Comment string `json:"-"`
}

// makeCode .
func makeCode(stmt *ast.CreateTableStmt, opt options) (string, []string, error) {
	importPath := make([]string, 0, 1)
	data := tmplData{
		TableName:    stmt.Table.Name.String(),
		RawTableName: stmt.Table.Name.String(),
		Fields:       make([]tmplField, 0, 1),
	}
	tablePrefix := opt.TablePrefix
	if tablePrefix != "" && strings.HasPrefix(data.TableName, tablePrefix) {
		data.NameFunc = true
		tableNameTemp := data.TableName[len(tablePrefix):]
		for _, v := range tableNameTemp {
			if !unicode.IsNumber(v) {
				data.TableName = data.TableName[len(tablePrefix):]
			}
			break
		}
	}
	if opt.ForceTableName || data.RawTableName != inflection.Plural(data.RawTableName) {
		data.NameFunc = true
	}

	data.TableName = toCamel(data.TableName)

	// find table comment
	for _, opt := range stmt.Options {
		if opt.Tp == ast.TableOptionComment {
			data.Comment = data.TableName + " " + opt.StrValue
			break
		}
	}

	if data.Comment == "" {
		data.Comment = data.TableName + "  ."
	}

	isPrimaryKey := make(map[string]bool)
	for _, con := range stmt.Constraints {
		if con.Tp == ast.ConstraintPrimaryKey {
			isPrimaryKey[con.Keys[0].Column.String()] = true
		}
	}

	columnPrefix := opt.ColumnPrefix
	for _, col := range stmt.Cols {
		colName := col.Name.Name.String()
		goFieldName := colName

		if columnPrefix != "" && strings.HasPrefix(goFieldName, columnPrefix) {
			goFieldName = goFieldName[len(columnPrefix):]
		}

		name := strings.Replace(toCamel(goFieldName), "id", "ID", -1)
		field := tmplField{
			Name: strings.Replace(name, "Url", "URL", -1),
		}

		tags := make([]string, 0, 4)
		// make GORM's tag
		gormTag := strings.Builder{}
		gormTag.WriteString("column:")
		gormTag.WriteString(colName)
		if opt.GormType {
			gormTag.WriteString(";type:")
			gormTag.WriteString(col.Tp.InfoSchemaStr())
		}
		if isPrimaryKey[colName] {
			gormTag.WriteString(";primary_key")
		}
		isNotNull := false
		canNull := false

		if opt.JSONTag {
			tags = append(tags, "json", colName)
		}

		for _, o := range col.Options {
			switch o.Tp {
			case ast.ColumnOptionPrimaryKey:
				if !isPrimaryKey[colName] {
					gormTag.WriteString(";primary_key")
					isPrimaryKey[colName] = true
				}
			case ast.ColumnOptionNotNull:
				isNotNull = true
			case ast.ColumnOptionAutoIncrement:
				gormTag.WriteString(";AUTO_INCREMENT")
			case ast.ColumnOptionDefaultValue:
				if value := getDefaultValue(o.Expr); value != "" {
					gormTag.WriteString(";default:")
					gormTag.WriteString(value)
				}
			case ast.ColumnOptionUniqKey:
				gormTag.WriteString(";unique")
			case ast.ColumnOptionNull:
				//gormTag.WriteString(";NULL")
				canNull = true
			case ast.ColumnOptionOnUpdate: // For Timestamp and Datetime only.
			case ast.ColumnOptionFulltext:
			case ast.ColumnOptionComment:
				field.Comment = strings.Replace(o.Expr.GetDatum().GetString(), "\n", "", -1)
			default:
				//return "", nil, errors.Errorf(" unsupport option %d\n", o.Tp)
			}
		}
		if !isPrimaryKey[colName] && isNotNull {
			gormTag.WriteString(";NOT NULL")
		}
		tags = append(tags, "gorm", gormTag.String())

		field.Tag = makeTagStr(tags)

		// get type in golang
		styleNull := opt.NullStyle
		if !canNull {
			styleNull = NullDisable
		}
		goType, pkg := mysqlToGoType(col.Tp, styleNull)
		if pkg != "" {
			importPath = append(importPath, pkg)
		}
		field.GoType = goType

		data.Fields = append(data.Fields, field)
	}

	builder := strings.Builder{}
	err := structTmpl.Execute(&builder, data)
	if err != nil {
		return "", nil, err
	}
	code, err := format.Source([]byte(builder.String()))
	if err != nil {
		return string(code), importPath, errors.WithMessage(err, "format golang code error")
	}
	return string(code), importPath, nil
}

// mysqlToGoType .
func mysqlToGoType(colTp *types.FieldType, style NullStyle) (name string, path string) {
	if style == NullInSQL {
		path = "database/sql"
		path = "database/sql"
		switch colTp.Tp {
		case mysql.TypeTiny, mysql.TypeShort, mysql.TypeInt24, mysql.TypeLong:
			name = "sql.NullInt32"
		case mysql.TypeLonglong:
			name = "sql.NullInt64"
		case mysql.TypeFloat, mysql.TypeDouble, mysql.TypeDecimal, mysql.TypeNewDecimal:
			name = "sql.NullFloat64"
		case mysql.TypeString, mysql.TypeVarchar, mysql.TypeVarString,
			mysql.TypeBlob, mysql.TypeTinyBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob:
			name = "sql.NullString"
		case mysql.TypeTimestamp, mysql.TypeDatetime, mysql.TypeDate:
			name = "sql.NullTime"
		//case mysql.TypeDecimal, mysql.TypeNewDecimal:
		//	name = "sql.NullString"
		case mysql.TypeJSON:
			name = "sql.NullString"
		default:
			return "UnSupport", ""
		}
	} else {
		switch colTp.Tp {
		case mysql.TypeTiny, mysql.TypeShort, mysql.TypeInt24, mysql.TypeLong:
			/*if mysql.HasUnsignedFlag(colTp.Flag) {
				name = "uint"
			} else {
				name = "int"
			}*/
			name = "int"
		case mysql.TypeLonglong:
			/*if mysql.HasUnsignedFlag(colTp.Flag) {
				name = "uint64"
			} else {
				name = "int64"
			}*/
			name = "int64"
		case mysql.TypeFloat, mysql.TypeDouble:
			name = "float64"
		case mysql.TypeDecimal, mysql.TypeNewDecimal:
			path = "github.com/shopspring/decimal"
			name = "decimal.Decimal"
		case mysql.TypeString, mysql.TypeVarchar, mysql.TypeVarString,
			mysql.TypeBlob, mysql.TypeTinyBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob:
			name = "string"
		case mysql.TypeTimestamp, mysql.TypeDatetime, mysql.TypeDate:
			path = "time"
			name = "time.Time"
		//case mysql.TypeDecimal, mysql.TypeNewDecimal:
		//	name = "string"
		case mysql.TypeJSON:
			name = "string"
		default:
			return "UnSupport", ""
		}
		if style == NullInPointer {
			name = "*" + name
		}
	}
	return
}

func makeTagStr(tags []string) string {
	builder := strings.Builder{}
	for i := 0; i < len(tags)/2; i++ {
		builder.WriteString(tags[i*2])
		builder.WriteString(`:"`)
		builder.WriteString(tags[i*2+1])
		builder.WriteString(`" `)
	}
	if builder.Len() > 0 {
		return builder.String()[:builder.Len()-1]
	}
	return builder.String()
}

func getDefaultValue(expr ast.ExprNode) (value string) {
	if expr.GetDatum().Kind() != types.KindNull {
		value = fmt.Sprintf("%v", expr.GetDatum().GetValue())
	} else if expr.GetFlag() != ast.FlagConstant {
		if expr.GetFlag() == ast.FlagHasFunc {
			if funcExpr, ok := expr.(*ast.FuncCallExpr); ok {
				value = funcExpr.FnName.O
			}
		}
	}
	return
}

func toCamel(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	s += "."

	n := strings.Builder{}
	n.Grow(len(s))
	temp := strings.Builder{}
	temp.Grow(len(s))
	wordFirst := true
	for _, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if wordFirst && vIsLow {
			v -= 'a' - 'A'
		}

		if vIsCap || vIsLow {
			temp.WriteByte(v)
			wordFirst = false
		} else {
			isNum := v >= '0' && v <= '9'
			wordFirst = isNum || v == '_' || v == ' ' || v == '-' || v == '.'
			if temp.Len() > 0 && wordFirst {
				word := temp.String()
				upper := strings.ToUpper(word)
				if _, ok := acronym[upper]; ok {
					n.WriteString(upper)
				} else {
					n.WriteString(word)
				}
				temp.Reset()
			}
			if isNum {
				n.WriteByte(v)
			}
		}
	}
	return n.String()
}

func initTemplate() {
	tmplParseOnce.Do(func() {
		var err error
		structTmpl, err = template.New("goStruct").Parse(structTmplRaw)
		if err != nil {
			panic(err)
		}
		fileTmpl, err = template.New("goFile").Parse(fileTmplRaw)
		if err != nil {
			panic(err)
		}
	})
}

func init() {
	structTmplRaw = `
{{- if .Comment -}}
// {{.Comment}}
{{end -}}
type {{.TableName}} struct {
{{- range .Fields}}
	{{.Name}} {{.GoType}} {{if .Tag}}` + "`{{.Tag}}`" + `{{end}}{{if .Comment}} // {{.Comment}}{{end}}
{{- end}}
}
{{if .NameFunc}}
// TableName .
func (m *{{.TableName}}) TableName() string {
	return "{{.RawTableName}}"
}
{{end}}`
	fileTmplRaw = `package {{.Package}}
{{if .ImportPath}}
import (
	{{- range .ImportPath}}
	"{{.}}"
	{{- end}}
)
{{- end}}
{{range .StructCode}}
{{.}}{{end}}`
}
