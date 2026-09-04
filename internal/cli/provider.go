package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	providerlib "github.com/roshbhatia/go-utils/provider"
	"github.com/roshbhatia/specutil/internal/graph"
	"github.com/spf13/cobra"
)

func newProviderCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "provider",
		Short: "Inspect optional suggestion providers",
	}
	command.AddCommand(newProviderListCmd(), newProviderValidateCmd())
	return command
}

func newProviderListCmd() *cobra.Command {
	var asJSON bool
	command := &cobra.Command{
		Use:   "list",
		Short: "List discovered suggestion providers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			providers, err := graph.SuggestionProviders()
			if err != nil {
				return err
			}
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(providers)
			}
			if len(providers) == 0 {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "No suggestion providers are configured.")
				return err
			}
			for _, loaded := range providers {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", loaded.Manifest.Name, loaded.Manifest.Description)
			}
			return nil
		},
	}
	command.Flags().BoolVar(&asJSON, "json", false, "emit provider metadata as JSON")
	return command
}

func newProviderValidateCmd() *cobra.Command {
	var asJSON bool
	command := &cobra.Command{
		Use:   "validate [name]",
		Short: "Validate provider manifests and runtime commands",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providers, err := graph.SuggestionProviders()
			if err != nil {
				return err
			}
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			var reports []providerlib.ValidationReport
			for _, loaded := range providers {
				if name != "" && loaded.Manifest.Name != name {
					continue
				}
				reports = append(reports, (providerlib.Validator{}).Validate(loaded.Manifest, "."))
			}
			if name != "" && len(reports) == 0 {
				return fmt.Errorf("suggestion provider %q was not found", name)
			}
			if asJSON {
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(reports); err != nil {
					return err
				}
			} else {
				for _, report := range reports {
					for _, check := range report.Checks {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", report.Provider, check.Status, check.Kind, check.Target)
					}
				}
			}
			var failures []error
			for _, report := range reports {
				if !report.OK() {
					failures = append(failures, report.Error())
				}
			}
			return errors.Join(failures...)
		},
	}
	command.Flags().BoolVar(&asJSON, "json", false, "emit validation reports as JSON")
	return command
}
