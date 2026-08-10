package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/domehahn/skcr/v2/internal/skillmeta"
	"github.com/spf13/cobra"
)

func newContractCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "Inspect and compare behavioral contracts",
	}
	cmd.AddCommand(newContractDiffCommand())
	cmd.AddCommand(newContractDigestCommand())
	return cmd
}

func newContractDiffCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "diff <old-contract-or-skill-dir> <new-contract-or-skill-dir>",
		Short: "Show deterministic security-relevant contract changes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldContract, err := skillmeta.LoadContract(resolveContractArgument(args[0]))
			if err != nil {
				return err
			}
			newContract, err := skillmeta.LoadContract(resolveContractArgument(args[1]))
			if err != nil {
				return err
			}
			if errs := skillmeta.ValidateContract(oldContract); len(errs) > 0 {
				return fmt.Errorf("old contract invalid: %s", strings.Join(errs, "; "))
			}
			if errs := skillmeta.ValidateContract(newContract); len(errs) > 0 {
				return fmt.Errorf("new contract invalid: %s", strings.Join(errs, "; "))
			}
			diff := skillmeta.DiffContracts(oldContract, newContract)
			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(diff)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Contract change: %s\n", diff.Classification)
			for _, change := range diff.Changes {
				subject := change.Capability
				if subject == "" {
					subject = change.Tool
				}
				if subject == "" {
					subject = change.Value
				}
				if change.Scope != "" {
					subject += ": " + change.Scope
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", strings.ToUpper(change.Impact), change.Type, subject)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output machine-readable JSON")
	return cmd
}

func newContractDigestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "digest <contract-or-skill-dir>",
		Short: "Print the SHA-256 digest of the normalized contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			contract, err := skillmeta.LoadContract(resolveContractArgument(args[0]))
			if err != nil {
				return err
			}
			if errs := skillmeta.ValidateContract(contract); len(errs) > 0 {
				return fmt.Errorf("contract invalid: %s", strings.Join(errs, "; "))
			}
			digest, err := skillmeta.ContractDigest(contract)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), digest)
			return nil
		},
	}
	return cmd
}

func resolveContractArgument(path string) string {
	if filepath.Base(path) == "contract.yaml" {
		return path
	}
	return filepath.Join(path, "contract.yaml")
}
