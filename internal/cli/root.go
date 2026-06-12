// Package cli wires the cobra command tree. The verb surface is deliberately
// small and deterministic: render, plan, diff, lock, graph, tui, serve. There is
// no `sync` verb — orchestration of remote writes lives in the shipped skills,
// never in the binary.
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/roshbhatia/specutil/internal/graph"
	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/provider/openspec"
	"github.com/roshbhatia/specutil/internal/render"
	"github.com/spf13/cobra"
)

// errNotImplemented is returned by verbs whose behavior lands in a later slice.
// It keeps the verb surface stable and discoverable while implementation fills in.
func notImplemented(verb string) error {
	return fmt.Errorf("%s: not implemented yet", verb)
}

// NewRootCmd builds the specutil root command and registers every verb.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "specutil",
		Short:         "Project OpenSpec change artifacts into other artifacts and visualizations",
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
		Use:   "plan",
		Short: "Emit a deterministic create/update/orphan plan for a sync target",
		RunE:  func(cmd *cobra.Command, args []string) error { return notImplemented("plan") },
	}
	cmd.Flags().String("target", "", "sync target: linear|notion")
	cmd.Flags().String("change", "", "change name to plan")
	return cmd
}

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Compare the local IR against the per-change lockfile",
		RunE:  func(cmd *cobra.Command, args []string) error { return notImplemented("diff") },
	}
}

func newLockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Read and write the CLI-managed identity map (content hash -> external ID)",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "get",
			Short: "Read a lock entry",
			RunE:  func(cmd *cobra.Command, args []string) error { return notImplemented("lock get") },
		},
		&cobra.Command{
			Use:   "set",
			Short: "Write a lock entry",
			RunE:  func(cmd *cobra.Command, args []string) error { return notImplemented("lock set") },
		},
	)
	return cmd
}

func newGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Project the cross-change dependency DAG (json|mermaid|dot)",
		Args:  cobra.NoArgs,
		RunE:  runGraph,
	}
	cmd.Flags().String("as", "json", "output format: json|mermaid|dot")
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
		RunE:  func(cmd *cobra.Command, args []string) error { return notImplemented("tui") },
	}
}

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve a static web site rendering the dependency DAG",
		RunE:  func(cmd *cobra.Command, args []string) error { return notImplemented("serve") },
	}
}
