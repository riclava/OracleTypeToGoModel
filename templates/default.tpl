package {{.PackageName}}

{{- if .ImportTime }}
import "time"
{{- end }}

type {{.Name}} struct {
{{- range .Metas}}
  {{.Name}} {{.Type}} `json:"{{.JsonName}}"`
{{- end}}
}

func (m *{{.Name}}) TableName() string {
  return "{{.TableName}}"
}
