package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/skcr/internal/platforms"
	"github.com/spf13/cobra"
)

func newCompatibilityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compatibility",
		Short: "Manage platform compatibility evidence",
	}
	cmd.AddCommand(newCompatibilityMatrixCommand())
	cmd.AddCommand(newCompatibilitySetCommand())
	cmd.AddCommand(newCompatibilityCheckCommand())
	return cmd
}

func newCompatibilityMatrixCommand() *cobra.Command {
	var target string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "matrix",
		Short: "Print platform compatibility matrix",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			matrix, err := platforms.LoadMatrix(root)
			if err != nil {
				return err
			}
			if jsonOut {
				return writeCompatibilityJSON(cmd, matrix)
			}
			for _, entry := range matrix {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\t%s\n", entry.Name, entry.MinVersion, entry.Status, entry.Validated, entry.Evidence)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func newCompatibilitySetCommand() *cobra.Command {
	var target string
	var minVersion string
	var evidence string
	var validated string
	var notes string
	cmd := &cobra.Command{
		Use:   "set <platform>",
		Short: "Set verified compatibility evidence for a platform",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			platform, err := models.NormalizePlatform(args[0])
			if err != nil {
				return err
			}
			entry := platforms.CompatibilityEntry{
				Name:       platform,
				MinVersion: minVersion,
				Status:     "verified",
				Source:     "local-evidence",
				Evidence:   evidence,
				Validated:  validated,
				Notes:      notes,
			}
			if err := platforms.SaveEvidence(root, entry); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "verified\t%s\t%s\t%s\n", platform, minVersion, evidence)
			return nil
		},
	}
	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	cmd.Flags().StringVar(&minVersion, "min-version", "", "Minimum validated platform version")
	cmd.Flags().StringVar(&evidence, "evidence", "", "Evidence file path relative to target")
	cmd.Flags().StringVar(&validated, "validated", "", "Validation date in YYYY-MM-DD format")
	cmd.Flags().StringVar(&notes, "notes", "", "Optional compatibility notes")
	_ = cmd.MarkFlagRequired("min-version")
	_ = cmd.MarkFlagRequired("evidence")
	_ = cmd.MarkFlagRequired("validated")
	return cmd
}

func newCompatibilityCheckCommand() *cobra.Command {
	var target string
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validate local compatibility evidence",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			matrix, err := platforms.LoadMatrix(root)
			if err != nil {
				return err
			}
			var errs []string
			verified := 0
			for _, entry := range matrix {
				if entry.Status == "verified" || entry.MinVersion != "unknown" {
					if valErr := platforms.ValidateEvidenceEntry(root, entry); valErr != nil {
						errs = append(errs, fmt.Sprintf("  %s: %v", entry.Name, valErr))
					} else {
						verified++
						fmt.Fprintf(cmd.OutOrStdout(), "ok\t%s\t%s\t%s\n", entry.Name, entry.MinVersion, entry.Evidence)
					}
				}
			}
			if len(errs) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%d error(s):\n", len(errs))
				for _, e := range errs {
					fmt.Fprintln(cmd.OutOrStdout(), e)
				}
				return fmt.Errorf("compatibility evidence has %d invalid entr%s", len(errs), map[bool]string{true: "y", false: "ies"}[len(errs) == 1])
			}
			if verified == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No verified entries found — all platforms are unverified")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d verified entr%s valid\n", verified, map[bool]string{true: "y", false: "ies"}[verified == 1])
			return nil
		},
	}
	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	return cmd
}

func writeCompatibilityJSON(cmd *cobra.Command, value any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
