package cli

import (
	"fmt"
	"path/filepath"

	"github.com/agentic-template-kit/skcr/internal/validator"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate generated/project state",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			errors, err := validator.ValidateProject(absTarget)
			if err != nil {
				return err
			}
			if len(errors) > 0 {
				for _, e := range errors {
					fmt.Println("ERROR", e)
				}
				return fmt.Errorf("validation failed")
			}
			fmt.Println("Validation passed")
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	return cmd
}
