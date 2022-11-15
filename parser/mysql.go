package parser

import (
	"database/sql" // "database/sql"

	_ "github.com/go-sql-driver/mysql" // "mysql"
	"github.com/pkg/errors"
)

// GetCreateTableFromDB .
func GetCreateTableFromDB(dsn, tableName string) (string, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", errors.WithMessage(err, "open db error")
	}
	defer db.Close()
	rows, err := db.Query("SHOW CREATE TABLE " + tableName)
	if err != nil {
		return "", errors.WithMessage(err, "query show create table error")
	}
	defer rows.Close()
	if !rows.Next() {
		return "", errors.Errorf("table(%s) not found", tableName)
	}
	var table string
	var createSQL string
	err = rows.Scan(&table, &createSQL)
	if err != nil {
		return "", err
	}
	return createSQL, nil
}

// getCreateTables .
func getCreateTablesByConfig(dsn string) ([]string, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.WithMessage(err, "open db error")
	}
	defer db.Close()
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, errors.WithMessage(err, "query show  tables error")
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			// handle this error
			panic(err)
		}
		tables = append(tables, table)
	}
	return tables, nil
}

// ParseSQLFromDB .
func ParseSQLFromDB(dsn, tableName string, options ...Option) (ModelCodes, error) {
	createSQL, err := GetCreateTableFromDB(dsn, tableName)
	if err != nil {
		return ModelCodes{}, err
	}
	return ParseSQL(createSQL, options...)
}

//GetCreateTables 获取待生成model的所有表
func GetCreateTables(dsn string, mysqlTable string) ([]string, error) {
	if mysqlTable != "*" {
		return []string{mysqlTable}, nil
	}

	tables, err := getCreateTablesByConfig(dsn)
	if err != nil {
		return nil, err
	}

	return tables, nil
}
