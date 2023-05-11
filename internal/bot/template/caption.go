package template

import "text/template"

var ApproveCaptions = template.Must(template.New("approve_captions").Parse(`
Подписей осталось: {{ .reviewed_count }} / {{ .total_disapproved_remained }}

text: {{ .text }}
author_id: {{ .author_id }}
created_at: {{ .created_at }}
`))
