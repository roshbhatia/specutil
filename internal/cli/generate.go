package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roshbhatia/go-utils/completion"
	sharedconfig "github.com/roshbhatia/go-utils/config"
	providerlib "github.com/roshbhatia/go-utils/provider"
	"github.com/roshbhatia/specutil/internal/graph"
	"github.com/roshbhatia/specutil/internal/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newCompletionCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:       "completion bash|zsh|fish|nu",
		Short:     "Generate shell completion",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "nu"},
		RunE: func(cmd *cobra.Command, args []string) error {
			generated, err := completion.Generate(args[0], completionMetadata(root))
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), generated)
			return err
		},
	}
}

func newValuesCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "__values changes|providers",
		Hidden: true,
		Args:   cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "providers" {
				for _, name := range graph.SuggestionProviderNames() {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), name)
				}
				return nil
			}
			if args[0] != "changes" {
				return fmt.Errorf("unknown completion value set %q", args[0])
			}
			repo := "."
			if len(args) == 2 {
				repo = repositoryFromContext(args[1])
			}
			provider, err := registry.SelectProvider(repo)
			if err != nil {
				return nil
			}
			names, err := provider.List()
			if err != nil {
				return nil
			}
			for _, name := range names {
				if !strings.ContainsAny(name, "\r\n") {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), name)
				}
			}
			return nil
		},
	}
}

func repositoryFromContext(context string) string {
	words := shellContextWords(context)
	repo := "."
	for index := 0; index < len(words); index++ {
		word := strings.Trim(words[index], `"'`)
		switch {
		case word == "--repo" || word == "-C":
			if index+1 < len(words) {
				repo = strings.Trim(words[index+1], `"'`)
				index++
			}
		case strings.HasPrefix(word, "--repo="):
			repo = strings.Trim(strings.TrimPrefix(word, "--repo="), `"'`)
		}
	}
	return repo
}

func shellContextWords(context string) []string {
	var words []string
	var current strings.Builder
	quote := rune(0)
	escaped := false
	flush := func() {
		if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}
	for _, one := range context {
		if escaped {
			current.WriteRune(one)
			escaped = false
			continue
		}
		if one == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if one == quote {
				quote = 0
			} else {
				current.WriteRune(one)
			}
			continue
		}
		if one == '\'' || one == '"' {
			quote = one
			continue
		}
		if one == ' ' || one == '\t' || one == '\n' {
			flush()
			continue
		}
		current.WriteRune(one)
	}
	if escaped {
		current.WriteRune('\\')
	}
	flush()
	return words
}

func newGenerateCmd(root *cobra.Command) *cobra.Command {
	var check bool
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Regenerate artifacts in the specutil source repository",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			return generateArtifacts(root, check)
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "fail when a generated artifact is stale")
	return cmd
}

func newConfigCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "Inspect project configuration support",
	}
	command.AddCommand(&cobra.Command{
		Use:   "schema",
		Short: "Print the project configuration JSON Schema",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			schema, err := projectConfigSchema()
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(schema)
			return err
		},
	})
	return command
}

func generateArtifacts(root *cobra.Command, check bool) error {
	schema, err := projectConfigSchema()
	if err != nil {
		return err
	}
	providerSchema, err := providerlib.Schema()
	if err != nil {
		return err
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		return fmt.Errorf("read README.md: %w", err)
	}
	generatedReadme, err := completion.ReplaceSection(
		string(readme),
		"cli",
		completion.Markdown(completionMetadata(root)),
	)
	if err != nil {
		return err
	}
	outputs := map[string][]byte{
		"README.md":                   []byte(generatedReadme),
		"schema/provider.schema.json": providerSchema,
		"schema/specutil.schema.json": schema,
	}
	for shell, path := range map[string]string{
		"bash": "completions/specutil.bash",
		"fish": "completions/specutil.fish",
		"nu":   "completions/specutil.nu",
		"zsh":  "completions/specutil.zsh",
	} {
		generated, generateErr := completion.Generate(shell, completionMetadata(root))
		if generateErr != nil {
			return generateErr
		}
		outputs[path] = []byte(generated + "\n")
	}
	for path, contents := range outputs {
		if check {
			current, readErr := os.ReadFile(path)
			if readErr != nil || string(current) != string(contents) {
				return fmt.Errorf("%s is stale; run specutil generate", path)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, contents, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func projectConfigSchema() ([]byte, error) {
	raw, err := sharedconfig.Schema[projectConfigDocument]("specutil project configuration")
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, fmt.Errorf("decode generated project schema: %w", err)
	}
	definitions, _ := document["$defs"].(map[string]any)
	for _, definition := range definitions {
		rule, _ := definition.(map[string]any)
		properties, _ := rule["properties"].(map[string]any)
		if _, hasID := properties["id"]; hasID {
			if _, hasSeverity := properties["severity"]; hasSeverity {
				rule["additionalProperties"] = true
			}
		}
	}
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode project schema: %w", err)
	}
	return append(encoded, '\n'), nil
}

type projectConfigDocument struct {
	Changes map[string]projectChangeConfig `json:"changes,omitempty"`
	Edges   []projectEdgeConfig            `json:"edges,omitempty"`
	Extract projectExtractConfig           `json:"extract,omitempty"`
	Check   projectCheckConfig             `json:"check,omitempty"`
}

type projectChangeConfig struct {
	DependsOn []string `json:"depends_on,omitempty"`
}

type projectEdgeConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type projectExtractConfig struct {
	Preset  string                 `json:"preset,omitempty"`
	Markers []projectExtractMarker `json:"markers,omitempty"`
	Fields  []projectExtractField  `json:"fields,omitempty"`
}

type projectExtractMarker struct {
	Key    string `json:"key"`
	Scope  string `json:"scope"  jsonschema:"enum=phase,enum=task,enum=scenario,enum=requirement"`
	Bullet string `json:"bullet"`
}

type projectExtractField struct {
	Key   string `json:"key"`
	Scope string `json:"scope" jsonschema:"enum=phase,enum=task,enum=scenario,enum=requirement"`
	Label string `json:"label"`
	Type  string `json:"type"  jsonschema:"enum=string,enum=list,enum=taskRefs"`
}

type projectCheckConfig struct {
	Preset  string             `json:"preset,omitempty"`
	Rules   []projectCheckRule `json:"rules,omitempty"`
	Disable []string           `json:"disable,omitempty"`
}

type projectCheckRule struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Severity string `json:"severity,omitempty" jsonschema:"enum=error,enum=warn"`
}

func completionMetadata(command *cobra.Command) completion.Command {
	return cobraCompletionMetadata(command, command.Name(), nil)
}

func cobraCompletionMetadata(command *cobra.Command, executable string, path []string) completion.Command {
	metadata := completion.Command{
		Name:            command.Name(),
		Synopsis:        command.Short,
		LongDescription: command.Long,
	}
	addCompletionFlags(&metadata, command)
	commandPath := strings.Join(path, " ")
	if acceptsChange(commandPath) {
		metadata.CompletionCommand = []string{executable, "__values", "changes", completion.ContextPlaceholder}
	}
	if commandPath == "provider validate" {
		metadata.CompletionCommand = []string{executable, "__values", "providers"}
	}
	for _, child := range command.Commands() {
		if !child.IsAvailableCommand() || child.Hidden || child.Name() == "completion" {
			continue
		}
		childPath := append(append([]string{}, path...), child.Name())
		metadata.Subcommands = append(metadata.Subcommands, cobraCompletionMetadata(child, executable, childPath))
	}
	return metadata
}

func addCompletionFlags(metadata *completion.Command, command *cobra.Command) {
	visit := func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		one := completion.Flag{
			Name:        flag.Name,
			Short:       flag.Shorthand,
			Description: flag.Usage,
			Value:       flag.NoOptDefVal == "",
		}
		if flag.Name == "change" {
			one.CompletionCommand = []string{"specutil", "__values", "changes", completion.ContextPlaceholder}
		}
		if flag.Name == "provider" {
			one.CompletionCommand = []string{"specutil", "__values", "providers"}
		}
		metadata.Flags = append(metadata.Flags, one)
	}
	command.NonInheritedFlags().VisitAll(visit)
	command.InheritedFlags().VisitAll(visit)
}

func acceptsChange(path string) bool {
	switch path {
	case "render", "check", "next", "review show", "review diff", "review set":
		return true
	default:
		return false
	}
}
