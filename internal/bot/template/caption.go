package template

import "text/template"

var ApproveCaptions = template.Must(template.New("approve_captions").Parse(`
Подписей осталось: {{ .reviewed_count }} / {{ .total_disapproved_remained }}

Текст: {{ .text }}
Автор: {{ .author_id }}
Дата создания: {{ .created_at }}
`))
