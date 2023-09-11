package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/riclava/oracletypeconverter/pkg/config"
	"github.com/riclava/oracletypeconverter/pkg/logger"
	"github.com/riclava/oracletypeconverter/pkg/oracle"
	"github.com/riclava/oracletypeconverter/pkg/tpl"
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
	ColumnName string
	Name       string
	JsonName   string
	Type       string
}

type Model struct {
	PackageName string
	Filename    string
	Name        string
	TableName   string
	ImportTime  bool
	ImportSql   bool
	Metas       []ModelMeta
	// create
	CombinedFields        string
	CombinedValues        string
	CombinedFieldsOnTable string
	// retrieve
	CombinedRetrieveKeys   string
	CombinedRetrieveFields string
	// update
	CombinedUpdateFields        string
	CombinedUpdateFieldsOnTable string
}

func fatal(err error) {
	logger.Fatalf("%v", err.Error())
}

func runConvert() {
	conf := config.AutoLoadConfig()

	err := os.MkdirAll(conf.ModelPath, 0755)
	if err != nil {
		fatal(err)
	}

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

	if len(tables) == 0 {
		logger.Warnf("No table found at current accout, program exit")
		return
	}

	logger.Infof("Found %v tables, start process all tables", len(tables))

	for idx, table := range tables {
		logger.Infof("Process table [%v, %v] ...", idx, table)

		if table != "FIELD_EXAMPLE" {
			continue
		}

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
			PackageName: conf.PackageName,
			Filename:    fmt.Sprintf("%s.go", path.Join(conf.ModelPath, strings.ToLower(underscoreToLowerCamel(metas[0].TableName)))),
			ImportTime:  false,
			Name:        underscoreToUpperCamel(strings.ToLower(metas[0].TableName)),
			TableName:   table,
		}

		for _, meta := range metas {
			jsonName := ""
			if conf.UpperCaseJson {
				jsonName = underscoreToUpperCamel(meta.ColumnName)
			} else {
				jsonName = underscoreToLowerCamel(meta.ColumnName)
			}
			mm := ModelMeta{
				ColumnName: meta.ColumnName,
				Name:       strings.ToUpper(meta.ColumnName),
				JsonName:   jsonName,
				Type:       "",
			}

			unsupported := false

			if isFieldString(meta.DataType) {
				if meta.Nullable == "Y" {
					mm.Type = "sql.NullString"
				} else {
					mm.Type = "string"
				}
			} else if isFloat64(meta.DataType, meta.DataPrecision, meta.DataScale) {
				if meta.Nullable == "Y" {
					mm.Type = "sql.NullFloat64"
				} else {
					mm.Type = "float64"
				}
			} else if isFloat32(meta.DataType) {
				if meta.Nullable == "Y" {
					mm.Type = "sql.NullFloat64"
				} else {
					mm.Type = "float32"
				}
			} else if isInt64(meta.DataType, meta.DataPrecision, meta.DataScale) {
				if meta.Nullable == "Y" {
					mm.Type = "sql.NullInt64"
				} else {
					mm.Type = "int64"
				}
			} else if isBytes(meta.DataType) {
				mm.Type = "[]byte"
			} else if isTime(meta.DataType) {
				if meta.Nullable == "Y" {
					mm.Type = "sql.NullTime"
				} else {
					model.ImportTime = true
					mm.Type = "time.Time"
				}

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

		// calc fields

		combinedFields := []string{}
		combinedValues := []string{}
		combinedFieldsOnTable := []string{}
		combinedRetrieveKeys := []string{}
		combinedRetrieveFields := []string{}
		combinedUpdateFields := []string{}
		combinedUpdateFieldsOnTable := []string{}

		for _, field := range model.Metas {
			combinedFields = append(combinedFields, fmt.Sprintf("\"%s\"", field.ColumnName))
			combinedValues = append(combinedValues, fmt.Sprintf(":%s", field.ColumnName))
			combinedFieldsOnTable = append(combinedFieldsOnTable, fmt.Sprintf("table.%s", field.Name))
			combinedRetrieveKeys = append(combinedRetrieveKeys, fmt.Sprintf("\"t\".\"%s\"", field.ColumnName))
			combinedRetrieveFields = append(combinedRetrieveFields, fmt.Sprintf("&r.%s", field.Name))
			combinedUpdateFields = append(combinedUpdateFields, fmt.Sprintf("\"%s\" = :%s", field.ColumnName, field.ColumnName))
			combinedUpdateFieldsOnTable = append(combinedUpdateFieldsOnTable, fmt.Sprintf("table.%s", field.Name))
		}

		model.CombinedFields = strings.Join(combinedFields, ",")
		model.CombinedValues = strings.Join(combinedValues, ",")
		model.CombinedFieldsOnTable = strings.Join(combinedFieldsOnTable, ",")
		model.CombinedRetrieveKeys = strings.Join(combinedRetrieveKeys, ",")
		model.CombinedRetrieveFields = strings.Join(combinedRetrieveFields, ",")
		model.CombinedUpdateFields = strings.Join(combinedUpdateFields, ",")
		model.CombinedUpdateFieldsOnTable = strings.Join(combinedUpdateFieldsOnTable, ",")

		if conf.ImportSql {
			model.ImportSql = true
		}

		templateStr := tpl.GetByFilename(conf.TemplateName)
		tmpl, err := template.New(conf.TemplateName).Funcs(template.FuncMap{
			"raw": func(s string) template.HTML {
				return template.HTML(s)
			},
		}).Parse(templateStr)
		if err != nil {
			fatal(err)
		}

		file, err := os.Create(model.Filename)
		if err != nil {
			fatal(err)
		}
		defer file.Close()

		err = tmpl.Execute(file, model)
		if err != nil {
			fatal(err)
		}
		logger.Infof("========================================")
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

func underscoreToUpperCamel(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		words[i] = strings.Title(word)
	}
	return strings.Join(words, "")
}

func underscoreToLowerCamel(s string) string {
	words := strings.Split(s, "_")
	for i := 1; i < len(words); i++ {
		words[i] = strings.Title(words[i])
	}
	return strings.Join(words, "")
}
