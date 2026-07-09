package render

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/roshbhatia/specutil/internal/ir"
)

//go:embed templates/*.tmpl
var embeddedTemplates embed.FS

// placeholder marks a mapped target section whose IR source was absent. The
// section is still emitted (so the canonical skeleton stays intact) but flagged.
const placeholder = "_None provided._"

// Options controls a render invocation.
type Options struct {
	// OverrideDir, if non-empty, is searched for `<target>.md.tmpl` before the
	// embedded default. A missing file there falls back to embedded with a warning.
	OverrideDir string
	// Title overrides the document title; defaults to the change name.
	Title string
}

// templateData is the value passed to the Go template.
type templateData struct {
	Title   string
	Change  *ir.Change
	Section map[string]string
}

// Render projects change into the named target, returning the rendered bytes and
// any warnings (e.g. absent source sections, override fallback). An unknown
// target is an error naming the supported set.
func Render(change *ir.Change, target string, opts Options) ([]byte, []ir.Warning, error) {
	mapping, ok := mappings[target]
	if !ok || internalTargets[target] {
		return nil, nil, fmt.Errorf("unknown render target %q; supported targets: %s",
			target, strings.Join(SupportedTargets(), ", "))
	}

	var warns []ir.Warning
	sections := make(map[string]string, len(mapping.Fields))
	for _, f := range mapping.Fields {
		content := strings.TrimSpace(f.Source(change))
		if content == "" {
			warns = append(warns, ir.Warning{
				File: change.Name,
				Msg:  fmt.Sprintf("render %s: source for section %q is absent; emitting placeholder", target, f.Key),
			})
			content = placeholder
		}
		sections[f.Key] = content
	}

	tmplText, tmplWarn, err := loadTemplate(target, opts.OverrideDir)
	if err != nil {
		return nil, warns, err
	}
	if tmplWarn != nil {
		warns = append(warns, *tmplWarn)
	}

	// Merge Sprig functions; our section func takes precedence on any conflict.
	funcs := sprig.FuncMap()
	funcs["section"] = func(m map[string]string, key string) string { return m[key] }
	tmpl, err := template.New(target).Funcs(funcs).Parse(tmplText)
	if err != nil {
		return nil, warns, fmt.Errorf("parsing %s template: %w", target, err)
	}

	title := opts.Title
	if title == "" {
		title = change.Name
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData{Title: title, Change: change, Section: sections}); err != nil {
		return nil, warns, fmt.Errorf("executing %s template: %w", target, err)
	}
	return buf.Bytes(), warns, nil
}

// IssueBodyData carries per-task fields injected alongside the change when
// rendering a github-issues body. These supplement the change-level sections.
type IssueBodyData struct {
	PhaseName string
	TaskRef   string
	TaskTitle string
}

// RenderIssueBody renders the github-issues body template for a single task.
// overrideDir follows the same semantics as Options.OverrideDir.
// The second return value carries the override-not-found warning when applicable.
func RenderIssueBody(change *ir.Change, d IssueBodyData, overrideDir string) (string, *ir.Warning, error) {
	const target = "github-issues"
	tmplText, tmplWarn, err := loadTemplate(target, overrideDir)
	if err != nil {
		return "", nil, err
	}

	sections := map[string]string{
		"summary": func() string {
			if change.Proposal != nil && change.Proposal.Why != "" {
				return change.Proposal.Why
			}
			return placeholder
		}(),
		"phase": d.PhaseName,
		"ref":   d.TaskRef,
		"title": d.TaskTitle,
	}

	funcs := sprig.FuncMap()
	funcs["section"] = func(m map[string]string, key string) string { return m[key] }
	tmpl, err := template.New(target).Funcs(funcs).Parse(tmplText)
	if err != nil {
		return "", nil, fmt.Errorf("parsing github-issues template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData{
		Title:   d.TaskTitle,
		Change:  change,
		Section: sections,
	}); err != nil {
		return "", nil, fmt.Errorf("executing github-issues template: %w", err)
	}
	return buf.String(), tmplWarn, nil
}

// loadTemplate returns the template text for target, preferring an override file
// and falling back to the embedded default with a warning if the override dir is
// set but lacks the file.
func loadTemplate(target, overrideDir string) (string, *ir.Warning, error) {
	name := target + ".md.tmpl"
	if overrideDir != "" {
		path := filepath.Join(overrideDir, name)
		if b, err := os.ReadFile(path); err == nil {
			return string(b), nil, nil
		}
		w := ir.Warning{Msg: fmt.Sprintf("override template %s not found; using embedded default", path)}
		b, err := embeddedTemplates.ReadFile("templates/" + name)
		if err != nil {
			return "", &w, fmt.Errorf("embedded template %s missing: %w", name, err)
		}
		return string(b), &w, nil
	}
	b, err := embeddedTemplates.ReadFile("templates/" + name)
	if err != nil {
		return "", nil, fmt.Errorf("embedded template %s missing: %w", name, err)
	}
	return string(b), nil, nil
}
