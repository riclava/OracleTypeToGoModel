package {{.PackageName}}

{{- if .ImportTime }}
import "time"
{{- end }}
{{- if .ImportSql }}
import "database/sql"
{{- end }}

type {{.Name}} struct {
{{- range .Metas}}
  {{.Name}} {{.Type}} `json:"{{.JsonName}}"`
{{- end}}
}

func (m *{{.Name}}) TableName() string {
  return "{{.TableName}}"
}

type {{.Name}}Service struct {
	db *sql.DB
}

func New{{.Name}}Service(db *sql.DB) *{{.Name}}Service {
	return &{{.Name}}Service{
		db: db,
	}
}

func (svc *{{.Name}}Service) TableName() string {
	item := {{.Name}}{}
	return item.TableName()
}

// Create
func (svc *{{.Name}}Service) Create(table *{{.Name}}) error {
	stmt, err := svc.db.Prepare(`INSERT INTO ` + svc.TableName() + `(` +
    `{{.CombinedFields | raw}}` +
		`) VALUES(` +
    `{{.CombinedValues | raw}}` +
		`)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	values := []interface{}{ {{.CombinedFieldsOnTable}} }

	if _, err := stmt.Exec(values...); err != nil {
		return err
	}

	return nil
}

// GetPage
func (svc *{{.Name}}Service) GetPage(page, size int, orderField string, order string, queryField, queryText string) ([]*{{.Name}}, int, error) {
	// wheres
	whereClause := ""
	if queryField != "" {
		whereClause += `WHERE ` + queryField + ` = :queryText`
	}

	countStmt, err := svc.db.Prepare(`SELECT COUNT(*) FROM ` + svc.TableName() + ` ` + whereClause)
	if err != nil {
		return nil, 0, err
	}
	defer countStmt.Close()

	var totalCount int
	if whereClause != "" {
		if err := countStmt.QueryRow(queryText).Scan(&totalCount); err != nil {
			return nil, 0, err
		}
	} else {
		if err := countStmt.QueryRow().Scan(&totalCount); err != nil {
			return nil, 0, err
		}
	}

	stmt, err := svc.db.Prepare(`
		SELECT *
		FROM (
				SELECT {{.CombinedRetrieveKeys | raw}}, ROWNUM AS rn
				FROM (
						SELECT * 
						FROM ` + svc.TableName() + ` ` +
						whereClause + ` ` +
						`ORDER BY "` + orderField + `" ` + order + `
				) "t"
				WHERE ROWNUM {{"<" | raw}}= :endId
		)
		WHERE rn > :startId
	`)
	if err != nil {
		return nil, 0, err
	}
	defer stmt.Close()

	// (startId, endId]
	startId := (page - 1) * size
	endId := startId + size

	values := []interface{}{}
	if whereClause != "" {
		values = append(values, queryText)
	}
	values = append(values, endId, startId)

	rows, err := stmt.Query(values...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var res []*{{.Name}}
	for rows.Next() {
		var r {{.Name}}
		var rn interface{}
		if err := rows.Scan({{.CombinedRetrieveFields | raw}}, &rn); err != nil {
			return nil, 0, err
		}
		res = append(res, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return res, totalCount, nil
}

// GetOne
func (svc *{{.Name}}Service) GetOne(keyField string, value interface{}) (*{{.Name}}, error) {
	stmt, err := svc.db.Prepare(`SELECT {{.CombinedRetrieveKeys | raw}} FROM ` + svc.TableName() + ` "t" WHERE "` + keyField + `" = :value AND ROWNUM=1`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(value)

	var r {{.Name}}
	if err := row.Scan({{.CombinedRetrieveFields | raw}}); err != nil {
		return nil, err
	}

	return &r, nil
}

// Update
func (svc *{{.Name}}Service) Update(keyField string, value interface{}, table *{{.Name}}) error {
	stmt, err := svc.db.Prepare("UPDATE " + svc.TableName() + " SET " +
		` {{.CombinedUpdateFields | raw}} ` +
		`WHERE "` + keyField + `" = :last`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	values := []interface{}{ {{.CombinedUpdateFieldsOnTable}} }

	values = append(values, value)

	if _, err := stmt.Exec(values...); err != nil {
		return err
	}

	return nil
}

// Delete
func (svc *{{.Name}}Service) Delete(keyField string, value interface{}) error {
	stmt, err := svc.db.Prepare(`DELETE FROM ` + svc.TableName() + ` WHERE "` + keyField + `" = :1 AND ROWNUM=1`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(value); err != nil {
		return err
	}

	return nil
}
