// Command agentsummons invokes agent harnesses headlessly from a single
// command surface.
package main

import (
	"errors"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is stamped by goreleaser on release builds (default ldflags,
// -X main.version). Must stay a package-level var named "version" for
// that stamping to land.
var version = "dev"

// cliVersion returns the stamped release version, falling back to the
// module version Go embeds under `go install` (where no ldflags run).
func cliVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		var ee *exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "agentsummons",
		Short: "Invoke agent harnesses headlessly",
		Long: `agentsummons invokes agent harnesses (Antigravity CLI, Claude Code,
Codex CLI) in headless mode: it owns the per-harness flag knowledge —
prompt passing, working-directory mechanics, capability syntax — so
callers and scripts don't have to rediscover it.

It knows nothing about the transcripts a harness writes; finding and
parsing those is the companion tool agentminutes' job.`,
		Version:      cliVersion(),
		SilenceUsage: true,
	}
	root.AddCommand(newBuildCmd(), newDoctorCmd(), newInfoCmd(), newRunCmd())
	return root
}
