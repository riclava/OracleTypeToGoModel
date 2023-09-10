package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/riclava/oracletypeconverter/pkg/config"
	"github.com/riclava/oracletypeconverter/pkg/logger"
	"github.com/riclava/oracletypeconverter/pkg/oracle"
)

type TableMeta struct {
	TableName     string
	ColumnId      sql.NullInt64
	ColumnName    string
	DataType      string
	DataLength    sql.NullInt64
	DataPrecision sql.NullInt64
	DataScale     sql.NullInt64
	Nullable      string
}

type ModelMeta struct {
	Name     string
	JsonName string
	Type     string
}
type Model struct {
	Name  string
	Metas []ModelMeta
}

func fatal(err error) {
	logger.Fatalf("%v", err.Error())
}

func runConvert() {
	conf := config.AutoLoadConfig()

	db, err := oracle.NewOracle(conf)
	if err != nil {
		logger.Fatalf("Error create oracle client instance: %v", err)
	}

	var tables []string
	rows, err := db.Query(`SELECT table_name FROM user_tables`)
	if err != nil {
		logger.Fatalf("Error list tables of current account: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		err := rows.Scan(
			&table,
		)
		if err != nil {
			fatal(err)
		}
		tables = append(tables, table)
	}
	err = rows.Err()
	if err != nil {
		fatal(err)
	}

	bits, err := json.MarshalIndent(tables, "", "  ")
	if err != nil {
		fatal(err)
	}

	logger.Infof(string(bits))
	if len(tables) == 0 {
		logger.Warnf("No table found at current accout, program exit")
		return
	}

	logger.Infof("Found %v tables, start process all tables", len(tables))

	for idx, table := range tables {
		logger.Infof("Process table %v ...", idx)
		rows, err := db.Query(`SELECT 
														table_name, 
														column_id, 
														column_name, 
														data_type, 
														data_length, 
														data_precision, 
														data_scale, 
														nullable
													FROM
														all_tab_columns
													WHERE table_name = :tableName
		`, table)
		if err != nil {
			fatal(err)
		}

		defer rows.Close()

		var metas []TableMeta
		for rows.Next() {
			var meta TableMeta
			err := rows.Scan(
				&meta.TableName,
				&meta.ColumnId,
				&meta.ColumnName,
				&meta.DataType,
				&meta.DataLength,
				&meta.DataPrecision,
				&meta.DataScale,
				&meta.Nullable,
			)
			if err != nil {
				fatal(err)
			}
			metas = append(metas, meta)
		}
		err = rows.Err()
		if err != nil {
			fatal(err)
		}

		var modelMetas []ModelMeta
		model := &Model{
			Name: metas[0].TableName,
		}

		for jdx, meta := range metas {
			logger.Infof("Process table %v, field[%v]: %v ...", table, jdx, meta.ColumnName)
			mm := ModelMeta{
				Name:     toPascalCase(meta.ColumnName),
				JsonName: toCamelCase(meta.ColumnName),
				Type:     "",
			}

			unsupported := false

			if isFieldString(meta.DataType) {
				mm.Type = "string"
			} else if isFloat64(meta.DataType, meta.DataPrecision, meta.DataScale) {
				mm.Type = "float64"
			} else if isFloat32(meta.DataType) {
				mm.Type = "float32"
			} else if isInt64(meta.DataType, meta.DataPrecision, meta.DataScale) {
				mm.Type = "int64"
			} else if isBytes(meta.DataType) {
				mm.Type = "[]byte"
			} else if isTime(meta.DataType) {
				mm.Type = "time.Time"
			} else {
				unsupported = true
				logger.Warnf("Table %v has unsupported data type %v with name %v", table, meta.DataType, meta.ColumnName)
			}

			if !conf.IgnoreUnsupportType && unsupported {
				fatal(errors.New("stop convert due to unsupported field"))
			}

			if !unsupported {
				modelMetas = append(modelMetas, mm)
			}
		}

		model.Metas = modelMetas

		bits, err := json.MarshalIndent(model, "", "  ")
		if err != nil {
			fatal(err)
		}

		logger.Infof("Table: %v \n%v\n", table, string(bits))
	}
}

func isFieldString(str string) bool {
	return str == "VARCHAR2" ||
		str == "NVARCHAR2" ||
		str == "LONG" ||
		str == "ROWID" ||
		str == "UROWID" ||
		str == "CHAR" ||
		str == "NCHAR" ||
		str == "CLOB" ||
		str == "NCLOB"
}

func isFloat64(str string, precision sql.NullInt64, scale sql.NullInt64) bool {
	if str == "FLOAT" || str == "BINARY_DOUBLE" {
		return true
	}

	// number 情况下，优先判断是否为 float64，如果不是，那么 number 直接看作 int64
	if str != "NUMBER" {
		return false
	}
	if precision.Valid && scale.Valid {
		if scale.Int64 == 0 && precision.Int64 <= 9 {
			return false
		} else {
			return true
		}
	} else if precision.Valid && !scale.Valid {
		if precision.Int64 <= 9 {
			return false
		} else {
			return true
		}
	}
	return false
}

func isInt64(str string, precision sql.NullInt64, scale sql.NullInt64) bool {
	return str == "NUMBER"
}

func isFloat32(str string) bool {
	return str == "BINARY_FLOAT"
}

func isBytes(str string) bool {
	return strings.Contains(str, "RAW") || str == "BLOB"
}

func isTime(str string) bool {
	return str == "DATE" || strings.HasPrefix(str, "TIMESTAMP")
}

// 转大驼峰（PascalCase）：首字母大写，单词首字母大写
func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})
	result := ""
	for _, word := range words {
		result += strings.Title(word)
	}
	return result
}

// 转小驼峰（camelCase）：首字母小写，单词首字母大写
func toCamelCase(s string) string {
	pascalCase := toPascalCase(s)
	return strings.ToLower(pascalCase[:1]) + pascalCase[1:]
}
