package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/roshbhatia/specutil/internal/ir"
)

func SupportedFormats() []string {
	return []string{"dot", "json", "mermaid"}
}

func (g *Graph) Project(format string) ([]byte, error) {
	switch format {
	case "json":
		return g.json()
	case "mermaid":
		return g.mermaid()
	case "dot":
		return g.dot()
	default:
		return nil, fmt.Errorf("unknown graph format %q; supported formats: %s",
			format, strings.Join(SupportedFormats(), ", "))
	}
}

func (g *Graph) json() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(g); err != nil {
		return nil, fmt.Errorf("encoding graph json: %w", err)
	}
	return buf.Bytes(), nil
}

var graphTemplateFunctions = template.FuncMap{
	"id": mermaidID,
	"quote": func(value string) string {
		return fmt.Sprintf("%q", value)
	},
}

var mermaidTemplate = template.Must(template.New("mermaid").Funcs(graphTemplateFunctions).Parse(`graph TD
{{range .Nodes}}  {{id .ID}}[{{quote .Label}}]
{{end}}{{range .Edges}}  {{id .From}} --> {{id .To}}
{{end}}`))

var dotTemplate = template.Must(template.New("dot").Funcs(graphTemplateFunctions).Parse(`digraph specutil {
  rankdir=LR;
{{range .Nodes}}  {{quote .ID}} [label={{quote .Label}}];
{{end}}{{range .Edges}}  {{quote .From}} -> {{quote .To}};
{{end}}}
`))

func (g *Graph) mermaid() ([]byte, error) {
	var b bytes.Buffer
	if err := mermaidTemplate.Execute(&b, g); err != nil {
		return nil, fmt.Errorf("render mermaid graph: %w", err)
	}
	return b.Bytes(), nil
}

func (g *Graph) dot() ([]byte, error) {
	var b bytes.Buffer
	if err := dotTemplate.Execute(&b, g); err != nil {
		return nil, fmt.Errorf("render dot graph: %w", err)
	}
	return b.Bytes(), nil
}

func mermaidID(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

type SuggestReport struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Capability string `json:"capability"`
}

func Suggest(changes []*ir.Change) []Candidate {
	type owner struct{ adds, mods []string }
	byCap := make(map[string]*owner)
	add := func(m map[string]*owner, capName, change string, isNew bool) {
		o := m[capName]
		if o == nil {
			o = &owner{}
			m[capName] = o
		}
		if isNew {
			o.adds = append(o.adds, change)
		} else {
			o.mods = append(o.mods, change)
		}
	}
	for _, c := range changes {
		if c.Proposal == nil {
			continue
		}
		for _, cap := range c.Proposal.Capabilities.New {
			add(byCap, cap.Name, c.Name, true)
		}
		for _, cap := range c.Proposal.Capabilities.Modified {
			add(byCap, cap.Name, c.Name, false)
		}
	}

	var out []Candidate
	caps := make([]string, 0, len(byCap))
	for name := range byCap {
		caps = append(caps, name)
	}
	sort.Strings(caps)
	for _, capName := range caps {
		o := byCap[capName]
		producers := append([]string(nil), o.adds...)
		sort.Strings(producers)
		consumers := append([]string(nil), o.mods...)
		sort.Strings(consumers)
		for _, p := range producers {
			for _, cons := range consumers {
				if p == cons {
					continue
				}
				out = append(out, Candidate{From: p, To: cons, Capability: capName})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		if out[i].To != out[j].To {
			return out[i].To < out[j].To
		}
		return out[i].Capability < out[j].Capability
	})
	return out
}
