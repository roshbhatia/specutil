// Package cli wires the cobra command tree. The verb surface is deliberately
// small and deterministic: render, plan, diff, lock, graph, tui, serve. There is
// no `sync` verb — orchestration of remote writes lives in the shipped skills,
// never in the binary.
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/roshbhatia/specutil/internal/detail"
	"github.com/roshbhatia/specutil/internal/graph"
	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/provider/openspec"
	"github.com/roshbhatia/specutil/internal/render"
	"github.com/roshbhatia/specutil/internal/syncplan"
	"github.com/roshbhatia/specutil/internal/tui"
	"github.com/roshbhatia/specutil/internal/web"
	"github.com/spf13/cobra"
)

// ErrNoMapping reports that a `lock get` found no entry. main maps it to exit
// code 3 so callers can distinguish "absent" from other failures.
func IsNoMapping(err error) bool {
	_, ok := err.(errNoMapping)
	return ok
}

// errNotImplemented is returned by verbs whose behavior lands in a later slice.
// It keeps the verb surface stable and discoverable while implementation fills in.
func notImplemented(verb string) error {
	return fmt.Errorf("%s: not implemented yet", verb)
}

// NewRootCmd builds the specutil root command and registers every verb.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "specutil",
		Short: "Project OpenSpec change artifacts into other artifacts and visualizations",
		Long: "specutil parses spec-framework change artifacts (OpenSpec in v1) into a " +
			"normalized IR and projects them into RFCs, design docs, tickets, dependency " +
			"graphs, and visualizations. The binary is deterministic and performs no network " +
			"I/O; remote writes are delegated to an agent via the shipped sync skills.",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	// Global flag: which repository to read changes from.
	root.PersistentFlags().StringP("repo", "C", ".", "repository root containing the openspec/ directory")

	root.AddCommand(
		newRenderCmd(),
		newPlanCmd(),
		newDiffCmd(),
		newLockCmd(),
		newGraphCmd(),
		newTUICmd(),
		newServeCmd(),
	)
	return root
}

func newRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render [change]",
		Short: "Render a change's IR into another artifact (rfc|design|tickets)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runRender,
	}
	cmd.Flags().String("as", "", "target artifact: rfc|design|tickets")
	cmd.Flags().String("change", "", "change name to render (or pass as positional arg)")
	cmd.Flags().String("templates", "", "override template directory")
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

// resolveChange loads the change named by --change or the first positional arg
// from the repository rooted at --repo.
func resolveChange(cmd *cobra.Command, args []string) (*ir.Change, error) {
	repo, _ := cmd.Flags().GetString("repo")
	name, _ := cmd.Flags().GetString("change")
	if name == "" && len(args) > 0 {
		name = args[0]
	}
	p := openspec.New(repo)
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

// emitWarnings prints parse/render warnings to stderr; the binary stays silent
// on stdout so rendered output is clean and pipeable.
func emitWarnings(cmd *cobra.Command, warns []ir.Warning) {
	for _, w := range warns {
		loc := w.File
		if w.Line > 0 {
			loc = fmt.Sprintf("%s:%d", w.File, w.Line)
		}
		if loc != "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %s\n", loc, w.Msg)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", w.Msg)
		}
	}
}

func newPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [change]",
		Short: "Emit a deterministic create/update/orphan plan for a sync target",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPlan,
	}
	cmd.Flags().String("target", "", "sync target namespace (e.g. linear|notion)")
	cmd.Flags().String("change", "", "change name to plan (or pass as positional arg)")
	cmd.Flags().StringP("out", "o", "", "write output to a file instead of stdout")
	return cmd
}

func runPlan(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")
	if target == "" {
		return fmt.Errorf("plan: --target is required")
	}
	repo, _ := cmd.Flags().GetString("repo")
	change, err := resolveChange(cmd, args)
	if err != nil {
		return err
	}
	emitWarnings(cmd, change.Warnings)
	lock, err := syncplan.LoadLock(repo, change.Name)
	if err != nil {
		return err
	}
	plan := syncplan.BuildPlan(change, lock, target)
	out, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	return writeOut(cmd, append(out, '\n'))
}

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [change]",
		Short: "Compare the local IR against the per-change lockfile",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runDiff,
	}
	cmd.Flags().String("target", "", "sync target namespace (e.g. linear|notion)")
	cmd.Flags().String("change", "", "change name to diff (or pass as positional arg)")
	cmd.Flags().StringP("out", "o", "", "write output to a file instead of stdout")
	return cmd
}

func runDiff(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")
	if target == "" {
		return fmt.Errorf("diff: --target is required")
	}
	repo, _ := cmd.Flags().GetString("repo")
	change, err := resolveChange(cmd, args)
	if err != nil {
		return err
	}
	emitWarnings(cmd, change.Warnings)
	lock, err := syncplan.LoadLock(repo, change.Name)
	if err != nil {
		return err
	}
	d := syncplan.DiffChange(change, lock, target)
	out, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return writeOut(cmd, append(out, '\n'))
}

func newLockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Read and write the CLI-managed identity map (identity hash -> external ID)",
	}
	cmd.PersistentFlags().String("target", "", "sync target namespace (e.g. linear|notion)")
	cmd.PersistentFlags().String("change", "", "change owning the lockfile")

	get := &cobra.Command{
		Use:   "get <identity>",
		Short: "Read a lock entry; exits 3 if no mapping exists",
		Args:  cobra.ExactArgs(1),
		RunE:  runLockGet,
	}
	set := &cobra.Command{
		Use:   "set <identity> <external-id>",
		Short: "Write a lock entry",
		Args:  cobra.ExactArgs(2),
		RunE:  runLockSet,
	}
	set.Flags().String("content-hash", "", "content hash to record for drift detection")
	set.Flags().String("title", "", "item title to retain for fuzzy re-match")

	cmd.AddCommand(get, set)
	return cmd
}

// errNoMapping signals `lock get` found no entry; main maps it to exit code 3
// so callers can distinguish "absent" from other failures.
type errNoMapping struct{ identity string }

func (e errNoMapping) Error() string { return fmt.Sprintf("no mapping for identity %q", e.identity) }

func lockChangeName(cmd *cobra.Command) (string, error) {
	name, _ := cmd.Flags().GetString("change")
	if name != "" {
		return name, nil
	}
	repo, _ := cmd.Flags().GetString("repo")
	names, err := openspec.New(repo).List()
	if err != nil {
		return "", err
	}
	if len(names) == 1 {
		return names[0], nil
	}
	return "", fmt.Errorf("specify --change (found: %v)", names)
}

func runLockGet(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")
	if target == "" {
		return fmt.Errorf("lock get: --target is required")
	}
	repo, _ := cmd.Flags().GetString("repo")
	change, err := lockChangeName(cmd)
	if err != nil {
		return err
	}
	lock, err := syncplan.LoadLock(repo, change)
	if err != nil {
		return err
	}
	ref, ok := lock.Get(target, args[0])
	if !ok {
		return errNoMapping{args[0]}
	}
	fmt.Fprintln(cmd.OutOrStdout(), ref.ExternalID)
	return nil
}

func runLockSet(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")
	if target == "" {
		return fmt.Errorf("lock set: --target is required")
	}
	repo, _ := cmd.Flags().GetString("repo")
	change, err := lockChangeName(cmd)
	if err != nil {
		return err
	}
	lock, err := syncplan.LoadLock(repo, change)
	if err != nil {
		return err
	}
	contentHash, _ := cmd.Flags().GetString("content-hash")
	title, _ := cmd.Flags().GetString("title")
	lock.Set(target, args[0], syncplan.Ref{ExternalID: args[1], ContentHash: contentHash, Title: title})
	return lock.Save(repo, change)
}

func newGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Project the cross-change dependency DAG (json|mermaid|dot)",
		Args:  cobra.NoArgs,
		RunE:  runGraph,
	}
	cmd.Flags().String("as", "json", "output format: json|mermaid|dot|detail")
	cmd.Flags().Bool("suggest", false, "report inferred candidate edges without mutating the manifest")
	cmd.Flags().StringP("out", "o", "", "write output to a file instead of stdout")
	return cmd
}

func runGraph(cmd *cobra.Command, args []string) error {
	repo, _ := cmd.Flags().GetString("repo")
	p := openspec.New(repo)
	changes, err := p.LoadAll()
	if err != nil {
		return err
	}
	for _, c := range changes {
		emitWarnings(cmd, c.Warnings)
	}

	if suggest, _ := cmd.Flags().GetBool("suggest"); suggest {
		cands := graph.Suggest(changes)
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
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %s\n", d.Kind, d.Msg)
	}

	format, _ := cmd.Flags().GetString("as")
	// The detail feed is the per-change ticket projection the visualizers drill
	// into; it shares graph's loader but is its own renderer-independent schema.
	if format == "detail" {
		out, err := detail.Build(changes).JSON()
		if err != nil {
			return err
		}
		return writeOut(cmd, out)
	}
	out, err := g.Project(format)
	if err != nil {
		return err
	}
	return writeOut(cmd, out)
}

// writeOut sends bytes to --out if set, otherwise stdout.
func writeOut(cmd *cobra.Command, b []byte) error {
	if outPath, _ := cmd.Flags().GetString("out"); outPath != "" {
		return os.WriteFile(outPath, b, 0o644)
	}
	cmd.OutOrStdout().Write(b)
	return nil
}

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open the workstream kanban and dependency graph TUI",
		Args:  cobra.NoArgs,
		RunE:  runTUI,
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	repo, _ := cmd.Flags().GetString("repo")
	changes, err := openspec.New(repo).LoadAll()
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
	return tui.Run(changes, g, diags)
}

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Generate a self-contained static web page rendering the dependency DAG",
		Long: "Generate a single self-contained HTML file that renders the cross-change\n" +
			"dependency DAG with Cytoscape.js: directed arrows, a dagre layered layout,\n" +
			"lifecycle-colored nodes, and a click-through ticket drawer. The page embeds\n" +
			"its data and runtime, so it works offline from file:// with no server. The\n" +
			"binary performs no network I/O; open the produced file in a browser.\n\n" +
			"If the graph has no edges, seed cross-change dependencies first with\n" +
			"`specutil graph --suggest` and record them in openspec/specutil.yaml.",
		Args: cobra.NoArgs,
		RunE: runServe,
	}
	cmd.Flags().StringP("out", "o", "specutil-graph.html", "output HTML file path ('-' for stdout)")
	return cmd
}

func runServe(cmd *cobra.Command, args []string) error {
	repo, _ := cmd.Flags().GetString("repo")
	changes, err := openspec.New(repo).LoadAll()
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
	g, _ := graph.Build(changes, manifest)
	html, err := web.Render(g, detail.Build(changes))
	if err != nil {
		return err
	}

	outPath, _ := cmd.Flags().GetString("out")
	if outPath == "-" {
		cmd.OutOrStdout().Write(html)
		return nil
	}
	if err := os.WriteFile(outPath, html, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s — open it in a browser\n", outPath)
	return nil
}
