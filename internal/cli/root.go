package cli

import (
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/roshbhatia/specutil/internal/check"
	"github.com/roshbhatia/specutil/internal/detail"
	"github.com/roshbhatia/specutil/internal/graph"
	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/registry"
	"github.com/roshbhatia/specutil/internal/render"
	"github.com/roshbhatia/specutil/internal/web"
	"github.com/spf13/cobra"
)

func NewRootCmd(version ...string) *cobra.Command {
	v := "dev"
	if len(version) > 0 && version[0] != "" {
		v = version[0]
	}
	root := &cobra.Command{
		Use:     "specutil",
		Short:   "Project OpenSpec change artifacts into other artifacts and visualizations",
		Version: v,
		Long: "specutil parses spec-framework change artifacts (OpenSpec in v1) into a " +
			"normalized IR and projects them into RFCs, design docs, tickets, dependency " +
			"graphs, and visualizations. It performs no network I/O.",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	root.PersistentFlags().StringP("repo", "C", ".", "repository root containing the openspec/ directory")

	root.AddCommand(
		newRenderCmd(),
		newGraphCmd(),
		newCheckCmd(),
		newNextCmd(),
		newReviewCmd(),
		newWebCmd(),
		newProviderCmd(),
		newConfigCmd(),
		newCompletionCmd(root),
		newGenerateCmd(root),
		newValuesCmd(),
	)
	root.CompletionOptions.DisableDefaultCmd = true
	return root
}

func newRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render [change]",
		Short: "Render a change into a shareable document (rfc|design|tickets)",
		Long: `Projects an OpenSpec change into a human-readable document for sharing with
stakeholders. Three output formats are supported:

  rfc      — RFC-style proposal doc (Why, What Changes, Requirements)
  design   — Technical design doc (Context, Goals, Decisions, Rollout)
  tickets  — Flat task checklist suitable for copy-paste into a tracker

Output goes to stdout by default; use -o to write a file. Combine with
git hooks or CI to auto-generate docs on change commits, or run manually
when preparing a design review or sprint planning session.

Typical invocations:
  specutil render --as rfc --change my-change
  specutil render --as tickets -o tickets.md`,
		Args: cobra.MaximumNArgs(1),
		RunE: runRender,
	}
	cmd.Flags().String("as", "", "target format: rfc|design|tickets (required)")
	cmd.Flags().String("change", "", "change name to render (or pass as positional arg)")
	cmd.Flags().String("templates", "", "override built-in template directory")
	cmd.Flags().StringP("out", "o", "", "write output to a file instead of stdout")
	return cmd
}

func runRender(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("as")
	if target == "" {
		return fmt.Errorf("render: --as is required (one of: %v)", render.SupportedTargets())
	}
	change, err := resolveChange(cmd, args)
	if err != nil {
		return err
	}

	overrideDir, _ := cmd.Flags().GetString("templates")
	out, warns, err := render.Render(change, target, render.Options{OverrideDir: overrideDir})
	if err != nil {
		return err
	}
	emitWarnings(cmd, change.Warnings)
	emitWarnings(cmd, warns)

	return writeOut(cmd, out)
}

func resolveChange(cmd *cobra.Command, args []string) (*ir.Change, error) {
	repo, _ := cmd.Flags().GetString("repo")
	name, _ := cmd.Flags().GetString("change")
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	p, err := registry.SelectProvider(repo)
	if err != nil {
		return nil, err
	}

	if name == "" {
		names, err := p.List()
		if err != nil {
			return nil, err
		}
		switch len(names) {
		case 0:
			return nil, fmt.Errorf("no changes found under %s/openspec/changes", repo)
		case 1:
			name = names[0]
		default:
			return nil, fmt.Errorf("multiple changes found; specify one with --change: %v", names)
		}
	}
	return p.Load(name)
}

func emitWarnings(cmd *cobra.Command, warns []ir.Warning) {
	for _, w := range warns {
		loc := w.File
		if w.Line > 0 {
			loc = fmt.Sprintf("%s:%d", w.File, w.Line)
		}
		if loc != "" {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %s\n", loc, w.Msg)
		} else {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", w.Msg)
		}
	}
}

func loadAllChanges(cmd *cobra.Command) ([]*ir.Change, error) {
	repo, _ := cmd.Flags().GetString("repo")
	p, err := registry.SelectProvider(repo)
	if err != nil {
		return nil, err
	}
	return p.LoadAll()
}

func newGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Output the cross-change dependency graph in various formats",
		Long: `Projects the cross-change dependency DAG into a machine-readable format for
integration with external tools. Primarily used by the skills and CI scripts;
for interactive browsing use ` + "`specutil web`" + ` instead.

Output formats:
  json     — Full graph model (nodes + edges) as JSON [default]
  mermaid  — Mermaid graph definition for embedding in docs or GitHub
  dot      — Graphviz DOT format for rendering with graphviz tools
  detail   — Per-change ticket detail feed (same as DETAIL in web view)

The --suggest flag infers candidate edges from shared capabilities without
writing anything. Pair it with an installed suggestion provider for deeper
semantic analysis. The optional command provider can run any configured agent.

Typical invocations:
  specutil graph --as mermaid                      # insert into a doc or README
  specutil graph --suggest                         # discover implied edges
  specutil graph --suggest --provider command --command my-agent
  specutil graph --as json | jq                    # pipe to other tools`,
		Args: cobra.NoArgs,
		RunE: runGraph,
	}
	cmd.Flags().String("as", "json", "output format: json|mermaid|dot|detail")
	cmd.Flags().Bool("suggest", false, "infer candidate edges from shared capabilities (read-only)")
	cmd.Flags().String("provider", "", "external suggestion provider (default: heuristic only)")
	cmd.Flags().String("command", "", "executable passed to the optional command provider")
	cmd.Flags().StringP("out", "o", "", "write output to a file instead of stdout")
	return cmd
}

func runGraph(cmd *cobra.Command, args []string) error {
	repo, _ := cmd.Flags().GetString("repo")
	changes, err := loadAllChanges(cmd)
	if err != nil {
		return err
	}
	for _, c := range changes {
		emitWarnings(cmd, c.Warnings)
	}

	if suggest, _ := cmd.Flags().GetBool("suggest"); suggest {
		provider, _ := cmd.Flags().GetString("provider")
		command, _ := cmd.Flags().GetString("command")
		var cands []graph.Candidate
		if provider != "" || command != "" {
			if provider == "" {
				provider = "command"
			}
			if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "running %s suggestion provider...\n", provider); err != nil {
				return err
			}
			var herr error
			cands, herr = graph.ProviderSuggest(cmd.Context(), changes, provider, command, repo)
			if herr != nil {
				return fmt.Errorf("provider suggest: %w", herr)
			}
		} else {
			cands = graph.Suggest(changes)
		}
		if cands == nil {
			cands = []graph.Candidate{}
		}
		out, err := json.MarshalIndent(graph.SuggestReport{Candidates: cands}, "", "  ")
		if err != nil {
			return err
		}
		return writeOut(cmd, append(out, '\n'))
	}

	manifest, err := graph.LoadManifest(repo)
	if err != nil {
		return err
	}
	g, diags := graph.Build(changes, manifest)
	for _, d := range diags {
		if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %s\n", d.Kind, d.Msg); err != nil {
			return err
		}
	}

	format, _ := cmd.Flags().GetString("as")

	if format == "detail" {
		out, err := detail.BuildWith(changes, reviewOptions(repo, changes)).JSON()
		if err != nil {
			return err
		}
		return writeOut(cmd, out)
	}
	out, err := g.Project(format)
	if err != nil {
		return fmt.Errorf("%w (or: detail)", err)
	}
	return writeOut(cmd, out)
}

func writeOut(cmd *cobra.Command, b []byte) error {
	if outPath, _ := cmd.Flags().GetString("out"); outPath != "" {
		return os.WriteFile(outPath, b, 0o644)
	}
	_, err := cmd.OutOrStdout().Write(b)
	return err
}

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [change]",
		Short: "Validate changes against the rubric declared in specutil.yaml",
		Long: `Checks each change against a declared rubric and exits non-zero when any
rule is violated, so it works as a pre-commit hook or a CI gate.

Rules are generic and parameterized; the repository supplies the specifics
under ` + "`check:`" + ` in openspec/specutil.yaml. When that block is absent and
openspec/config.yaml names a schema specutil ships a preset for, that preset
applies automatically. A repository with neither is not checked.

Every rule reads only what the author stated: a heading that is present, a
marker that is declared, a bullet that follows another. None infers intent
from prose, so two runs over the same input always agree.

Exit codes:
  0  every rule passed (warnings may still be reported)
  1  at least one error-severity rule was violated

Typical invocations:
  specutil check                     # every change
  specutil check my-change           # one change
  specutil check --as json | jq      # machine-readable findings
  specutil check --list-rules        # what the resolved rubric enforces`,
		Args: cobra.MaximumNArgs(1),
		RunE: runCheck,
	}
	cmd.Flags().String("change", "", "check a single change (or pass as positional arg)")
	cmd.Flags().String("as", "text", "output format: text|json")
	cmd.Flags().Bool("list-rules", false, "list every built-in rule and exit")
	cmd.Flags().StringP("out", "o", "", "write output to a file instead of stdout")
	return cmd
}

func runCheck(cmd *cobra.Command, args []string) error {
	if list, _ := cmd.Flags().GetBool("list-rules"); list {
		var b strings.Builder
		for _, id := range check.RuleIDs() {
			fmt.Fprintf(&b, "%-26s %s\n", id, check.RuleDoc(id))
		}
		return writeOut(cmd, []byte(b.String()))
	}

	repo, _ := cmd.Flags().GetString("repo")
	name, _ := cmd.Flags().GetString("change")

	if len(args) > 0 && name == "" {
		if derivedRepo, derivedName, ok := changeDirTarget(args[0]); ok {
			repo, name = derivedRepo, derivedName
			args = nil
			if err := cmd.Flags().Set("repo", repo); err != nil {
				return err
			}
			if err := cmd.Flags().Set("change", name); err != nil {
				return err
			}
		}
	}

	manifest, err := graph.LoadManifest(repo)
	if err != nil {
		return err
	}
	cfg, err := manifest.CheckConfig(repo)
	if err != nil {
		return err
	}
	if cfg.IsZero() {
		if _, err := fmt.Fprintf(cmd.ErrOrStderr(),
			"no rubric declared: add a `check:` block to %s or name a known schema in openspec/config.yaml\n",
			graph.ManifestFile); err != nil {
			return err
		}
		return nil
	}

	var changes []*ir.Change
	if name != "" || len(args) > 0 {
		c, cerr := resolveChange(cmd, args)
		if cerr != nil {
			return cerr
		}
		changes = []*ir.Change{c}
	} else {
		changes, err = loadAllChanges(cmd)
		if err != nil {
			return err
		}
	}
	for _, c := range changes {
		emitWarnings(cmd, c.Warnings)
	}

	report, err := check.Run(cfg, changes)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("as")
	switch format {
	case "json":
		out, merr := json.MarshalIndent(report, "", "  ")
		if merr != nil {
			return merr
		}
		if werr := writeOut(cmd, append(out, '\n')); werr != nil {
			return werr
		}
	case "text":
		if werr := writeOut(cmd, []byte(checkText(report))); werr != nil {
			return werr
		}
	default:
		return fmt.Errorf("unknown check format %q; supported formats: json, text", format)
	}

	if !report.OK() {
		cmd.SilenceErrors = true
		return errCheckFailed
	}
	return nil
}

func changeDirTarget(arg string) (repo, name string, ok bool) {
	info, err := os.Stat(arg)
	if err != nil || !info.IsDir() {
		return "", "", false
	}
	abs, err := filepath.Abs(arg)
	if err != nil {
		return "", "", false
	}
	abs = filepath.Clean(abs)
	changesDir := filepath.Dir(abs)
	if filepath.Base(changesDir) != "changes" {
		return "", "", false
	}
	openspecDir := filepath.Dir(changesDir)
	if filepath.Base(openspecDir) != "openspec" {
		return "", "", false
	}
	return filepath.Dir(openspecDir), filepath.Base(abs), true
}

var errCheckFailed = errors.New("check: rubric violated")

type checkTextFinding struct {
	Severity string
	Rule     string
	Message  string
	Location string
}

type checkTextSection struct {
	Change   string
	Header   bool
	Findings []checkTextFinding
}

//go:embed check-output.txt.tmpl
var checkOutputSource string

var checkOutputTemplate = template.Must(
	template.New("check-output").Parse(checkOutputSource),
)

func checkText(r *check.Report) string {
	data := struct {
		Sections []checkTextSection
		OK       bool
		Count    int
		Noun     string
		Warnings int
		Errors   int
	}{
		OK:       r.OK(),
		Count:    len(r.Checked),
		Noun:     "changes",
		Warnings: r.Warnings(),
		Errors:   r.Errors(),
	}
	if data.Count == 1 {
		data.Noun = "change"
	}
	change := ""
	for index, finding := range r.Findings {
		if index == 0 || change != finding.Change {
			change = finding.Change
			data.Sections = append(data.Sections, checkTextSection{
				Change: finding.Change,
				Header: finding.Change != "" || index > 0,
			})
		}
		location := finding.File
		if finding.Line > 0 {
			location = fmt.Sprintf("%s:%d", finding.File, finding.Line)
		}
		if location != "" {
			location = " (" + location + ")"
		}
		section := &data.Sections[len(data.Sections)-1]
		section.Findings = append(section.Findings, checkTextFinding{
			Severity: string(finding.Severity),
			Rule:     finding.Rule,
			Message:  finding.Msg,
			Location: location,
		})
	}

	var rendered bytes.Buffer
	if err := checkOutputTemplate.Execute(&rendered, data); err != nil {
		panic(err)
	}
	return rendered.String()
}

func newWebCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Open a browser view of the change board, dependency graph, and task details",
		Long: `Renders all OpenSpec changes into a self-contained HTML file and opens it
in the default browser. The page has three views:

  Kanban  — lifecycle board (proposed / active / archived) with per-change
            progress meters. Click a card to open the detail drilldown.

  Graph   — dependency DAG laid out in waves. A wave is the set of changes
            whose prerequisites are all satisfied at the same depth, so every
            change in a wave can be worked in parallel. Node color encodes
            readiness (ready / in progress / blocked / waiting / done).

  Detail  — per-change drilldown: execution plan (stages → tasks), Why /
            What Changes narrative, outstanding tasks, and per-stage chart.

Every task takes a comment or a removal request, and the Detail view collects
a decision. 'Copy feedback' and 'Download' produce the JSON that
` + "`specutil review ingest`" + ` folds back into the change. Nothing is posted: there
is no server behind the page.

Pass --diff to review the working-tree code alongside the plan. It runs git
locally, needs --change to say which change the diff belongs to, and defaults
its base to the commit recorded at that change's last review.

A fresh file is written to the system temp directory on each invocation so
you always see current data; old files accumulate in /tmp and can be cleared
periodically. Pass -o to write a specific path or '-' for stdout.`,
		Args: cobra.NoArgs,
		RunE: runWeb,
	}
	cmd.Flags().StringP("out", "o", "", "output HTML file path (default: timestamped temp file; '-' for stdout)")
	cmd.Flags().Bool("open", true, "open the generated page in the default browser")
	cmd.Flags().Bool("diff", false, "include the working-tree diff for annotation (requires a single change)")
	cmd.Flags().String("change", "", "change the --diff belongs to")
	cmd.Flags().String("base", "", "git ref for --diff (default: the reviewed commit, else HEAD)")
	return cmd
}

func runWeb(cmd *cobra.Command, args []string) error {
	repo, _ := cmd.Flags().GetString("repo")
	changes, err := loadAllChanges(cmd)
	if err != nil {
		return err
	}
	for _, c := range changes {
		emitWarnings(cmd, c.Warnings)
	}
	manifest, err := graph.LoadManifest(repo)
	if err != nil {
		return err
	}
	g, diags := graph.Build(changes, manifest)

	opts := reviewOptions(repo, changes)
	if err := attachDiff(cmd, repo, changes, &opts); err != nil {
		return err
	}

	html, err := web.Render(g, detail.BuildWith(changes, opts), diags, graph.Suggest(changes))
	if err != nil {
		return err
	}

	outPath, _ := cmd.Flags().GetString("out")
	if outPath == "-" {
		_, err := cmd.OutOrStdout().Write(html)
		return err
	}
	if outPath == "" {
		var b [4]byte
		rand.Read(b[:])
		outPath = filepath.Join(os.TempDir(), "specutil-web-"+hex.EncodeToString(b[:])+".html")
	}
	if err := os.WriteFile(outPath, html, 0o644); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", outPath); err != nil {
		return err
	}

	if open, _ := cmd.Flags().GetBool("open"); open {
		if err := openInBrowser(outPath); err != nil {
			if _, writeErr := fmt.Fprintf(cmd.ErrOrStderr(), "could not open a browser (%v); open %s yourself\n", err, outPath); writeErr != nil {
				return writeErr
			}
		}
	}
	return nil
}

func openInBrowser(path string) error {
	var bin string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		bin = "open"
	case "windows":
		bin, args = "rundll32", []string{"url.dll,FileProtocolHandler"}
	default:
		bin = "xdg-open"
	}
	return exec.Command(bin, append(args, path)...).Start()
}
