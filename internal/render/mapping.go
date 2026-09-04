package render

import (
	"bytes"
	"sort"
	"strings"
	"text/template"

	"github.com/roshbhatia/specutil/internal/export"
	"github.com/roshbhatia/specutil/internal/ir"
)

type Field struct {
	Key string

	Source func(*ir.Change) string
}

type Mapping struct {
	Fields []Field
}

var mappings = map[string]Mapping{
	"rfc": {
		Fields: []Field{
			{"summary", proposalWhy},
			{"motivation", proposalWhatChanges},
			{"guide", guideLevel},
			{"reference", specsMarkdown},
			{"drawbacks", designRisks},
			{"alternatives", designDecisions},
			{"unresolved", designOpenQuestions},
			{"future", proposalNonGoals},
		},
	},
	"design": {
		Fields: []Field{
			{"context", designContext},
			{"goals", designGoals},
			{"decisions", designDecisions},
			{"risks", designRisks},
			{"migration", designMigration},
			{"openquestions", designOpenQuestions},
			{"proposal", proposalWhatChanges},
		},
	},
	"tickets": {
		Fields: []Field{
			{"summary", proposalWhy},
		},
	},
}

func SupportedTargets() []string {
	out := make([]string, 0, len(mappings))
	for k := range mappings {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func proposalWhy(c *ir.Change) string {
	if c.Proposal == nil {
		return ""
	}
	return c.Proposal.Why
}

func proposalWhatChanges(c *ir.Change) string {
	if c.Proposal == nil {
		return ""
	}
	return c.Proposal.WhatChanges
}

func proposalNonGoals(c *ir.Change) string {
	if c.Proposal == nil {
		return ""
	}
	return c.Proposal.NonGoals
}

func designContext(c *ir.Change) string {
	if c.Design == nil {
		return ""
	}
	return c.Design.Context
}

func designGoals(c *ir.Change) string {
	if c.Design == nil {
		return ""
	}
	return c.Design.Goals
}

func designDecisions(c *ir.Change) string {
	if c.Design == nil {
		return ""
	}
	return c.Design.Decisions
}

func designRisks(c *ir.Change) string {
	if c.Design == nil {
		return ""
	}
	return c.Design.Risks
}

func designMigration(c *ir.Change) string {
	if c.Design == nil {
		return ""
	}
	return c.Design.Migration
}

func designOpenQuestions(c *ir.Change) string {
	if c.Design == nil {
		return ""
	}
	return c.Design.OpenQuestions
}

func guideLevel(c *ir.Change) string {
	data := struct {
		Capabilities []ir.Capability
		Context      string
	}{}
	if c.Proposal != nil {
		data.Capabilities = append(append([]ir.Capability{}, c.Proposal.Capabilities.New...), c.Proposal.Capabilities.Modified...)
	}
	if c.Design != nil && c.Design.Context != "" {
		data.Context = c.Design.Context
	}
	var rendered bytes.Buffer
	if err := guideTemplate.Execute(&rendered, data); err != nil {
		panic(err)
	}
	return strings.TrimRight(rendered.String(), "\n")
}

var guideTemplate = template.Must(template.New("guide").Parse(`{{range .Capabilities}}- **{{.Name}}**{{if .Description}} — {{.Description}}{{end}}
{{end}}{{if and .Capabilities .Context}}
{{end}}{{.Context}}`))

type specStep struct {
	Keyword string
	Text    string
}

type specCriterion struct {
	Name  string
	Steps []specStep
}

type specGroup struct {
	Requirement string
	Text        string
	Criteria    []specCriterion
}

var specsTemplate = template.Must(template.New("specs").Parse(`{{range .}}#### {{.Requirement}}

{{if .Text}}{{.Text}}

{{end}}{{range .Criteria}}- **{{.Name}}**
{{range .Steps}}  - {{if .Keyword}}{{.Keyword}} {{end}}{{.Text}}
{{end}}{{end}}
{{end}}`))

func specsMarkdown(c *ir.Change) string {
	specs := append([]*ir.Spec{}, c.Specs...)
	sort.SliceStable(specs, func(i, j int) bool { return specs[i].Capability < specs[j].Capability })
	sorted := &ir.Change{Name: c.Name, Specs: specs}

	var groups []specGroup
	for _, group := range export.BuildChange(sorted).CriteriaByRequirement() {
		view := specGroup{
			Requirement: group.Requirement,
			Text:        requirementText(specs, group.Requirement),
		}
		for _, cr := range group.Criteria {
			criterion := specCriterion{Name: cr.Name}
			criterion.Steps = appendSteps(criterion.Steps, "Given", cr.Given)
			criterion.Steps = appendSteps(criterion.Steps, "When", cr.When)
			criterion.Steps = appendSteps(criterion.Steps, "Then", cr.Then)
			criterion.Steps = appendSteps(criterion.Steps, "", cr.Steps)
			view.Criteria = append(view.Criteria, criterion)
		}
		groups = append(groups, view)
	}
	var rendered bytes.Buffer
	if err := specsTemplate.Execute(&rendered, groups); err != nil {
		panic(err)
	}
	return strings.TrimRight(rendered.String(), "\n")
}

func appendSteps(target []specStep, keyword string, steps []string) []specStep {
	for _, text := range steps {
		target = append(target, specStep{Keyword: keyword, Text: text})
	}
	return target
}

func requirementText(specs []*ir.Spec, want string) string {
	for _, s := range specs {
		for _, r := range s.Requirements {
			if export.Humanize(r.Name) == want {
				return r.Text
			}
		}
	}
	return ""
}
