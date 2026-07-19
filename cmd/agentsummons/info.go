package main

import (
	"fmt"

	"github.com/agent-ecosystem/agentsummons"
	"github.com/spf13/cobra"
)

// infoEnvelope is the info --json output shape.
type infoEnvelope struct {
	SchemaVersion int                         `json:"schema_version"`
	Harnesses     []agentsummons.Capabilities `json:"harnesses"`
}

func newInfoCmd() *cobra.Command {
	var (
		harnessID string
		asJSON    bool
	)
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show each harness's capability manifest",
		Long: `Info reports what each harness supports and via which syntax: prompt
passing, working-directory mechanics, model pinning, session identity,
resume, tool allowlists, JSON output, permission bypass, and the quirk
notes. Descriptive only — for correctly assembled argv, use build.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ids := agentsummons.Harnesses()
			if harnessID != "" {
				ids = []agentsummons.ID{agentsummons.ID(harnessID)}
			}
			all := make([]agentsummons.Capabilities, 0, len(ids))
			for _, id := range ids {
				caps, err := agentsummons.InfoFor(id)
				if err != nil {
					return requestExit(cmd, err)
				}
				all = append(all, caps)
			}
			if asJSON {
				return writeJSON(cmd, infoEnvelope{SchemaVersion: schemaVersion, Harnesses: all})
			}
			for i, caps := range all {
				if i > 0 {
					_, _ = fmt.Fprintln(cmd.OutOrStdout())
				}
				printCapabilities(cmd, caps)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&harnessID, "harness", "", "limit to one harness: "+harnessHelpList())
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit the manifests as JSON")
	return cmd
}

func printCapabilities(cmd *cobra.Command, caps agentsummons.Capabilities) {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "%s (%s)\n", caps.Harness, caps.Binary)
	line := func(name, value string) {
		if value == "" {
			value = "(unsupported)"
		}
		_, _ = fmt.Fprintf(out, "  %-14s %s\n", name, value)
	}
	_, _ = fmt.Fprintf(out, "  %-14s %s %v\n", "version", caps.Binary, caps.VersionArgs)
	if len(caps.BaseArgs) > 0 {
		_, _ = fmt.Fprintf(out, "  %-14s %v\n", "base args", caps.BaseArgs)
	}
	line("prompt", caps.Prompt)
	line("workdir", caps.Workdir)
	line("model", caps.Model)
	line("session-id", caps.SessionID)
	line("resume", caps.Resume)
	line("allowed-tools", caps.AllowedTools)
	jsonOutput := caps.JSONOutput
	if jsonOutput != "" {
		jsonOutput += " (" + caps.JSONOutputShape + ")"
	}
	line("json-output", jsonOutput)
	line("auto-approve", caps.AutoApprove)
	for _, note := range caps.Notes {
		_, _ = fmt.Fprintf(out, "  note: %s\n", note)
	}
}
