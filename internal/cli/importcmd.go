package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newImportCommand() *cobra.Command {
	var target string
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "import <tarball>",
		Short: "Unpack a skill bundle tarball into the skill sources directory",
		Long: `Reads a .tar.gz bundle produced by 'skcr bundle' and extracts each skill
directory into skill_sources.output_dir.

Existing files are skipped unless --force is passed.
Use --dry-run to preview what would be extracted without writing anything.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tarball := args[0]

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			destDir := filepath.Join(abs, sourceBase)

			f, err := os.Open(tarball)
			if err != nil {
				return fmt.Errorf("open tarball: %w", err)
			}
			defer f.Close()

			gz, err := gzip.NewReader(f)
			if err != nil {
				return fmt.Errorf("open gzip: %w", err)
			}
			defer gz.Close()

			tr := tar.NewReader(gz)
			created, skipped := 0, 0

			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("read tarball: %w", err)
				}
				if hdr.Typeflag == tar.TypeDir {
					continue
				}

				dest := filepath.Join(destDir, filepath.FromSlash(hdr.Name))

				if _, err := os.Stat(dest); err == nil && !force {
					rel, _ := filepath.Rel(abs, dest)
					fmt.Fprintf(cmd.OutOrStdout(), "  skip  %s (already exists)\n", rel)
					skipped++
					continue
				}

				rel, _ := filepath.Rel(abs, dest)
				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "  would create  %s\n", rel)
					created++
					continue
				}

				if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
					return fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
				}
				out, err := os.Create(dest)
				if err != nil {
					return fmt.Errorf("create %s: %w", dest, err)
				}
				if _, err := io.Copy(out, tr); err != nil {
					out.Close()
					return fmt.Errorf("write %s: %w", dest, err)
				}
				out.Close()
				fmt.Fprintf(cmd.OutOrStdout(), "  +  %s\n", rel)
				created++
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\nDry run: would create %d file(s), skip %d.\n", created, skipped)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nImported %d file(s), skipped %d.\n", created, skipped)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}
