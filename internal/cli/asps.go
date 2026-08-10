package cli

import (
	"fmt"
	"path/filepath"

	"github.com/domehahn/skcr/v2/internal/asps"
	"github.com/domehahn/skcr/v2/internal/skillmeta"
	"github.com/spf13/cobra"
)

func newASPSCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "asps", Short: "Author ASPS v1.0 requirements without asserting compliance"}
	cmd.AddCommand(newASPSListCommand(), newASPSShowCommand(), newASPSValidateCommand(), newASPSRequirementsCommand(), newASPSCoverageCommand())
	return cmd
}

func newASPSListCommand() *cobra.Command {
	var domains, profiles, jsonOut bool
	cmd := &cobra.Command{Use: "list", Short: "List the pinned ASPS v1.0 authoring catalog", RunE: func(cmd *cobra.Command, _ []string) error {
		if profiles {
			values := asps.Profiles()
			if jsonOut {
				return writeJSON(cmd, values)
			}
			for _, value := range values {
				fmt.Fprintln(cmd.OutOrStdout(), value)
			}
			return nil
		}
		if domains {
			values := asps.Domains()
			if jsonOut {
				return writeJSON(cmd, values)
			}
			for _, value := range values {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", value.ID, value.Name)
			}
			return nil
		}
		values := asps.Properties()
		if jsonOut {
			return writeJSON(cmd, values)
		}
		for _, value := range values {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", value.ID, value.Name)
		}
		return nil
	}}
	cmd.Flags().BoolVar(&domains, "domains", false, "List domains instead of properties")
	cmd.Flags().BoolVar(&profiles, "profiles", false, "List skcr requirement profiles")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")
	return cmd
}

func newASPSShowCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{Use: "show <property-or-profile>", Short: "Show an ASPS property or requirement profile", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if asps.KnownProfile(args[0]) {
			value := map[string]any{"profile": args[0], "properties": asps.ProfileProperties(args[0])}
			if jsonOut {
				return writeJSON(cmd, value)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", args[0])
			for _, property := range asps.ProfileProperties(args[0]) {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", property)
			}
			return nil
		}
		property, ok := asps.FindProperty(args[0])
		if !ok {
			return fmt.Errorf("unknown ASPS property or profile %q", args[0])
		}
		if jsonOut {
			return writeJSON(cmd, property)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\nDomain:\t%s\nEvidence route:\t%s\n", property.ID, property.Name, property.DomainID, asps.EvidenceRoute(property.ID))
		return nil
	}}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")
	return cmd
}

func loadSkillAssurance(skillDir string) (skillmeta.Descriptor, skillmeta.AssuranceDocument, error) {
	descriptor, err := skillmeta.LoadDescriptor(skillmeta.DescriptorPath(skillDir))
	if err != nil {
		return skillmeta.Descriptor{}, skillmeta.AssuranceDocument{}, err
	}
	if descriptor.Assurance == nil {
		return descriptor, skillmeta.AssuranceDocument{}, fmt.Errorf("descriptor has no assurance reference")
	}
	document, err := skillmeta.LoadAssurance(filepath.Join(skillDir, descriptor.Assurance.File))
	return descriptor, document, err
}

func newASPSValidateCommand() *cobra.Command {
	return &cobra.Command{Use: "validate <skill>", Short: "Validate ASPS and assurance requirements", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if errs := skillmeta.ValidateDirectory(args[0]); len(errs) > 0 {
			return fmt.Errorf("source validation failed: %v", errs)
		}
		_, document, err := loadSkillAssurance(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "VALID ASPS %s (%d required properties)\n", document.ASPS.Version, len(skillmeta.RequiredASPSProperties(document)))
		return nil
	}}
}

func newASPSRequirementsCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{Use: "requirements <skill>", Short: "Expand profiles into the effective ASPS requirement set", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		_, document, err := loadSkillAssurance(args[0])
		if err != nil {
			return err
		}
		properties := skillmeta.RequiredASPSProperties(document)
		if jsonOut {
			return writeJSON(cmd, map[string]any{"profiles": document.ASPS.RequiredProfiles, "properties": properties})
		}
		for _, property := range properties {
			fmt.Fprintln(cmd.OutOrStdout(), property)
		}
		return nil
	}}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")
	return cmd
}

type aspsCoverageItem struct {
	Property       string `json:"property"`
	SourceEvidence string `json:"source_evidence"`
	Status         string `json:"status"`
}

func newASPSCoverageCommand() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{Use: "coverage <skill>", Short: "Report source-artifact coverage for required ASPS properties", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if errs := skillmeta.ValidateDirectory(args[0]); len(errs) > 0 {
			return fmt.Errorf("source validation failed: %v", errs)
		}
		_, document, err := loadSkillAssurance(args[0])
		if err != nil {
			return err
		}
		items := []aspsCoverageItem{}
		for _, property := range skillmeta.RequiredASPSProperties(document) {
			items = append(items, aspsCoverageItem{property, asps.EvidenceRoute(property), "routed"})
		}
		if jsonOut {
			return writeJSON(cmd, map[string]any{"meaning": "source declarations exist; this is not a PASS/FAIL assurance result", "coverage": items})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Source coverage only — verification results are owned by skil.")
		for _, item := range items {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", item.Status, item.Property, item.SourceEvidence)
		}
		return nil
	}}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON")
	return cmd
}
