package cli

import (
	"fmt"
	"path/filepath"

	"github.com/domehahn/skcr/internal/validator"
	"github.com/spf13/cobra"
)

var cliAbsPathValidate = filepath.Abs

func newValidateCommand() *cobra.Command {
	var target string
	var againstLock string
	var skills bool
	var platform string
	var ci bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate generated/project state",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := cliAbsPathValidate(target)
			if err != nil {
				return err
			}
			errors, err := validator.ValidateProjectWithOptions(absTarget, validator.Options{
				AgainstLock: againstLock,
				Skills:      skills,
				Platform:    platform,
				CI:          ci,
			})
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
	cmd.Flags().StringVar(&againstLock, "against-lock", "", "Validate against skpm agent-skills.lock")
	cmd.Flags().BoolVar(&skills, "skills", false, "Validate configured skpm skill lock state")
	cmd.Flags().StringVar(&platform, "platform", "", "Validate only the selected canonical platform")
	cmd.Flags().BoolVar(&ci, "ci", false, "Enable CI-oriented validation output")
	return cmd
}
