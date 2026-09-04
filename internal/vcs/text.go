package vcs

import (
	"bytes"
	"fmt"
	"text/template"
)

var diffTextTemplate = template.Must(template.New("diff-text").Funcs(template.FuncMap{
	"heading": func(file File) string {
		switch file.Status {
		case StatusRenamed:
			return fmt.Sprintf("%s (renamed from %s)", file.Path, file.OldPath)
		case StatusBinary:
			return fmt.Sprintf("%s (binary)", file.Path)
		default:
			return fmt.Sprintf("%s (%s)", file.Path, file.Status)
		}
	},
	"marker": marker,
}).Parse(`{{if .Diff.Note}}No diff: {{.Diff.Note}}
{{else if not .Files}}No changes against {{.Diff.Base}}.
{{else}}{{.Files}} {{.Noun}} changed against {{.Diff.Base}}: +{{.Added}} -{{.Deleted}}
{{range .Diff.Files}}
{{heading .}}
{{range .Hunks}}  {{.Header}}  [{{.Identity}}]
{{range .Lines}}  {{marker .Kind}}{{.Text}}
{{end}}{{end}}{{end}}{{end}}`))

func (d *Diff) Text() string {
	files, added, deleted := d.Stats()
	noun := "files"
	if files == 1 {
		noun = "file"
	}
	data := struct {
		Diff    *Diff
		Files   int
		Noun    string
		Added   int
		Deleted int
	}{Diff: d, Files: files, Noun: noun, Added: added, Deleted: deleted}

	var rendered bytes.Buffer
	if err := diffTextTemplate.Execute(&rendered, data); err != nil {
		panic(err)
	}
	return rendered.String()
}

func marker(kind string) string {
	switch kind {
	case LineAdd:
		return "+"
	case LineDelete:
		return "-"
	default:
		return " "
	}
}
