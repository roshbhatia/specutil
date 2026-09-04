package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/roshbhatia/specutil/internal/lifecycle"
	"github.com/spf13/cobra"
)

func newNextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next [change]",
		Short: "Report which subtasks are runnable now",
		Long: `Answers one question: what runs now.

A tasks.md declares a shape, a dependency edge per subtask, and a stop
condition. Without a consumer those declarations are documentation, and the
work gets done top to bottom whatever the graph says. This reads them.

Readiness never crosses a phase, because a phase is a boundary between runs.
The reported phase is the lowest-numbered one still holding pending work; the
ready set is every pending subtask in it whose dependencies are complete.

A graph phase with more than one runnable subtask reports concurrent, so the
caller can fan out. A loop phase does not: its next iteration reads the state
the current one writes. Owner gates and adversarial reviews are never counted
as fan-out work.

Exit codes:
  0  a ready set was reported, or every task is complete
  2  work remains but nothing is runnable, which means a dependency cycle

Typical invocations:
  specutil next                      # the active change
  specutil next my-change            # one change
  specutil next --as json | jq       # drive a runner from the ready set`,
		Args: cobra.MaximumNArgs(1),
		RunE: runNext,
	}
	cmd.Flags().String("change", "", "report a single change (or pass as positional arg)")
	cmd.Flags().String("as", "text", "output format: text|json")
	cmd.Flags().StringP("out", "o", "", "write output to a file instead of stdout")
	return cmd
}

func runNext(cmd *cobra.Command, args []string) error {
	c, err := resolveChange(cmd, args)
	if err != nil {
		return err
	}
	emitWarnings(cmd, c.Warnings)

	n := lifecycle.ComputeNext(c)

	format, _ := cmd.Flags().GetString("as")
	if format == "json" {
		body, merr := json.MarshalIndent(n, "", "  ")
		if merr != nil {
			return merr
		}
		if werr := writeOut(cmd, append(body, '\n')); werr != nil {
			return werr
		}
	} else if werr := writeOut(cmd, []byte(renderNext(n))); werr != nil {
		return werr
	}

	if !n.Done && len(n.Ready) == 0 {
		return errDependencyCycle{change: n.Change}
	}
	return nil
}

type errDependencyCycle struct{ change string }

func (e errDependencyCycle) Error() string {
	return fmt.Sprintf("%s: work remains but no subtask is runnable, so its dependencies form a cycle", e.change)
}

func IsDependencyCycle(err error) bool {
	_, ok := err.(errDependencyCycle)
	return ok
}

var nextTemplate = template.Must(template.New("next").Funcs(template.FuncMap{
	"firstWords": firstWords,
	"join":       strings.Join,
	"label":      label,
}).Parse(`{{if .Next.Done}}{{.Next.Change}}: every task is complete
{{else}}{{.Next.Change}}
phase {{.Next.Phase}}. {{.Next.PhaseName}} ({{.Shape}})
{{if .Next.Stop}}
stop: {{.Next.Stop}}
{{end}}
ready ({{len .Next.Ready}}){{if .Next.Concurrent}}, runnable concurrently{{else if .OrderUnstated}}, order unstated: the phase declares no ` + "`deps:`" + `, so run them in listed order{{end}}:
{{if .Next.Ready}}{{range .Next.Ready}}{{printf "  %-6s %-8s %s\n" .ID (label .) (firstWords .Text 14)}}{{end}}{{else}}  none
{{end}}{{if .Next.Blocked}}
blocked ({{len .Next.Blocked}}):
{{range .Next.Blocked}}{{printf "  %-6s waits on %s\n" .ID (join .WaitsOn ", ")}}{{end}}{{end}}{{end}}`))

func renderNext(n lifecycle.Next) string {
	shape := n.Shape
	if shape == "" {
		shape = "no shape declared"
	}
	data := struct {
		Next          lifecycle.Next
		Shape         string
		OrderUnstated bool
	}{
		Next:          n,
		Shape:         shape,
		OrderUnstated: n.Shape == "graph" && !n.EdgesDeclared && len(n.Ready) > 1,
	}
	var rendered bytes.Buffer
	if err := nextTemplate.Execute(&rendered, data); err != nil {
		panic(err)
	}
	return rendered.String()
}

func label(t lifecycle.Task) string {
	switch {
	case t.Adverse:
		return "review"
	case t.Gate:
		return t.Kind
	case t.Kind == "" || t.Kind == "plain":
		return "task"
	default:
		return t.Kind
	}
}

func firstWords(text string, n int) string {
	words := strings.Fields(text)
	if len(words) <= n {
		return strings.Join(words, " ")
	}
	return strings.Join(words[:n], " ") + " ..."
}
