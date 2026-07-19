package main

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

// schemaVersion versions the JSON output shapes of run, build, info, and
// doctor — a cross-language contract, reviewed on any shape change.
const schemaVersion = 1

// writeJSON emits one envelope, indented, on the command's stdout.
func writeJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
