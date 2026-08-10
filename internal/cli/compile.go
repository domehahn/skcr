package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/domehahn/skcr/v2/internal/compiler"
	"github.com/domehahn/skcr/v2/internal/skillmeta"
	"github.com/spf13/cobra"
)

func newCompileCommand() *cobra.Command {
	var target, output string
	var requireLossless bool
	cmd := &cobra.Command{
		Use:   "compile [skill] <name-or-path>",
		Short: "Compile source skills into executable assurance artifacts",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if target != compiler.Target {
				return fmt.Errorf("unsupported compile target %q (supported: skil)", target)
			}
			input := args[len(args)-1]
			if len(args) == 2 && args[0] != "skill" {
				return fmt.Errorf("expected 'compile skill <name>' or 'compile <path>'")
			}
			if len(args) == 2 {
				input = filepath.Join(".agents", "skills", input)
			}
			absInput, err := filepath.Abs(input)
			if err != nil {
				return err
			}
			absOutput := output
			if absOutput == "" {
				absOutput, err = filepath.Abs(filepath.Join(".skcr", "build", compiler.Target))
			} else {
				absOutput, err = filepath.Abs(absOutput)
			}
			if err != nil {
				return err
			}
			dirs, err := compileSkillDirs(absInput)
			if err != nil {
				return err
			}
			if len(dirs) == 0 {
				return fmt.Errorf("no source skills found under %s", input)
			}
			for _, dir := range dirs {
				result, err := compiler.CompileSkill(dir, compiler.Options{OutputRoot: absOutput, RequireLossless: requireLossless, CompilerVersion: Version})
				if err != nil {
					return fmt.Errorf("compile %s: %w", dir, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "compiled  %s  -> %s\n", result.SkillName, result.OutputDir)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&target, "target", compiler.Target, "Compilation target")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output root (default: .skcr/build/skil)")
	cmd.Flags().BoolVar(&requireLossless, "require-lossless", false, "Fail if any source semantics cannot be mapped or preserved")
	cmd.AddCommand(newCompileInspectCommand())
	return cmd
}

func compileSkillDirs(input string) ([]string, error) {
	info, err := os.Stat(input)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("compile input must be a skill or collection directory")
	}
	if _, err := os.Stat(skillmeta.DescriptorPath(input)); err == nil {
		return []string{input}, nil
	}
	entries, err := os.ReadDir(input)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(input, entry.Name())
		if _, err := os.Stat(skillmeta.DescriptorPath(dir)); err == nil {
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func newCompileInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use: "inspect <build-manifest>", Short: "Inspect a compiler build manifest", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			var manifest compiler.Manifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return fmt.Errorf("invalid build manifest: %w", err)
			}
			if manifest.SchemaVersion != "1" {
				return fmt.Errorf("unsupported build manifest schema_version %q", manifest.SchemaVersion)
			}
			formatted, _ := json.MarshalIndent(manifest, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(formatted))
			return nil
		},
	}
}
