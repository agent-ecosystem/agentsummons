package main

import (
	"github.com/agent-ecosystem/agentsummons"
	"github.com/spf13/cobra"
)

// buildEnvelope is the build output shape.
type buildEnvelope struct {
	SchemaVersion int      `json:"schema_version"`
	Harness       string   `json:"harness"`
	Argv          []string `json:"argv"`
	PromptIndex   int      `json:"prompt_index"`
	Dir           string   `json:"dir"`
	EnvAdded      []string `json:"env_added,omitempty"`
}

func newBuildCmd() *cobra.Command {
	var f requestFlags
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Assemble the command line without executing it",
		Long: `Build prints the exact argv, working directory, and environment
additions run would use, as JSON, without invoking anything.

This is the correct-assembly path for callers in other languages: ordering
quirks (like Antigravity's prompt-must-be-last) are implemented once, here,
instead of being reimplemented from the info manifest. prompt_index marks
the prompt's position in argv so archivers can redact it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := f.request()
			if err != nil {
				return err
			}
			built, err := agentsummons.Build(req)
			if err != nil {
				return requestExit(cmd, err)
			}
			return writeJSON(cmd, buildEnvelope{
				SchemaVersion: schemaVersion,
				Harness:       string(req.Harness),
				Argv:          built.Argv,
				PromptIndex:   built.PromptIndex,
				Dir:           built.Dir,
				EnvAdded:      built.EnvAdded,
			})
		},
	}
	addRequestFlags(cmd, &f)
	return cmd
}
