package review

import (
	"bytes"
	"strings"
	"text/template"
)

var markdownTemplate = template.Must(template.New("review-markdown").Funcs(template.FuncMap{
	"hunkLocation": func(h HunkStatus) string {
		if h.Header == "" {
			return h.File
		}
		return h.File + " " + oneLine(h.Header)
	},
	"lines": func(value string) []string {
		lines := strings.Split(strings.TrimSpace(value), "\n")
		for index := range lines {
			lines[index] = strings.TrimSpace(lines[index])
		}
		return lines
	},
	"oneLine":  oneLine,
	"shortSHA": shortSHA,
}).Parse(`{{define "item"}}- [{{.Phase}}] {{oneLine .Text}}
{{if .Comment}}{{range lines .Comment}}  > {{.}}
{{end}}{{end}}{{end}}# Review: {{.Status.Change}}

{{if not .Status.Reviewed}}Decision: none recorded. This change has not been reviewed.
{{else}}Decision: {{.Status.Decision}}
{{if .Status.Stale}}Status: stale. The artifacts changed after this decision (reviewed {{.Status.ReviewHash}}, now {{.Status.ChangeHash}}).
{{else}}Status: current. The artifacts match what was reviewed.
{{end}}{{if .Status.Note}}
## Note

{{.Status.Note}}
{{end}}{{if .ChangeComment}}
## Change comment

{{.ChangeComment}}
{{end}}{{if .Status.Dropped}}
## Requested removals

{{range .Status.Dropped}}{{template "item" .}}{{end}}{{end}}{{if .Comments}}
## Comments

{{range .Comments}}{{template "item" .}}{{end}}{{end}}{{if .Status.Hunks}}
## Code comments

{{range .Status.Hunks}}- {{hunkLocation .}}
{{range lines .Comment}}  > {{.}}
{{end}}{{end}}{{end}}{{if .Moved}}
## Changed since review

{{range .Moved}}- [{{.Phase}}] {{oneLine .Text}} ({{.Drift}})
{{end}}{{end}}{{if .Status.BaseCommit}}
Code reviewed from {{shortSHA .Status.BaseCommit}}. Run ` + "`specutil review diff {{.Status.Change}}`" + ` for what moved since.
{{end}}{{if .NoOpen}}
No open comments and no drift since the review.
{{end}}{{end}}`))

type markdownData struct {
	Status        *Status
	ChangeComment string
	Comments      []ItemStatus
	Moved         []ItemStatus
	NoOpen        bool
}

func Markdown(st *Status) string {
	data := markdownData{Status: st}
	for _, annotation := range st.Annotations {
		if annotation.Scope == ScopeChange && strings.TrimSpace(annotation.Comment) != "" {
			data.ChangeComment = annotation.Comment
			break
		}
	}
	for _, item := range st.Items {
		if item.Comment != "" && item.Action != ActionDrop {
			data.Comments = append(data.Comments, item)
		}
		if item.Drift == DriftNew || item.Drift == DriftChanged {
			data.Moved = append(data.Moved, item)
		}
	}
	data.NoOpen = len(st.Dropped) == 0 && len(data.Comments) == 0 && len(data.Moved) == 0 && len(st.Hunks) == 0

	var rendered bytes.Buffer
	if err := markdownTemplate.Execute(&rendered, data); err != nil {
		panic(err)
	}
	return rendered.String()
}

func shortSHA(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
