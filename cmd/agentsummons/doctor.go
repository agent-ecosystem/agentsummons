package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/agent-ecosystem/agentsummons"
	"github.com/spf13/cobra"
)

// doctorEnvelope is the doctor --json output shape.
type doctorEnvelope struct {
	SchemaVersion int            `json:"schema_version"`
	Harnesses     []doctorReport `json:"harnesses"`
}

type doctorReport struct {
	Harness   string `json:"harness"`
	Installed string `json:"installed,omitempty"`
	Validated string `json:"validated"`
	// Status is one of the status* values below.
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// doctorReport.Status values — part of the JSON contract.
const (
	statusClean        = "clean"           // installed at or below validated
	statusDrift        = "drift-candidate" // installed newer than validated
	statusNotInstalled = "not-installed"
	statusError        = "error" // version probe failed
)

// Doctor's exit codes; runDoctor keeps the worst one seen. Unlike run,
// doctor never propagates a harness exit code, so plain small values are
// unambiguous here.
const (
	exitDoctorDrift       = 1
	exitDoctorProbeFailed = 2
)

// versionProbeTimeout bounds each harness's version command; these are
// free, local invocations that should never take long.
const versionProbeTimeout = 30 * time.Second

func newDoctorCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Compare installed harness versions against the validated flag surface (free)",
		Long: `Doctor reports, for each supported harness, the installed version
against the newest release this build's flag knowledge was validated
against. Nothing is invoked beyond each harness's version command and
nothing is spent — it is the cheap pre-experiment gate.

A newer installed version is a drift candidate, not a failure: flags
usually survive releases, and a removed flag fails loudly at run time.

Exit codes: 0 clean (missing harnesses are fine), 1 at least one drift
candidate, 2 a version probe failed.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd, asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit the report as JSON")
	return cmd
}

func runDoctor(cmd *cobra.Command, asJSON bool) error {
	reports := make([]doctorReport, 0, len(agentsummons.Harnesses()))
	worst := 0
	for _, id := range agentsummons.Harnesses() {
		report := doctorReport{Harness: string(id), Validated: agentsummons.LastValidated[id]}
		ctx, cancel := context.WithTimeout(cmd.Context(), versionProbeTimeout)
		installed, err := agentsummons.Version(ctx, id)
		cancel()
		var nie *agentsummons.NotInstalledError
		switch {
		case errors.As(err, &nie):
			report.Status = statusNotInstalled
		case err != nil:
			report.Status = statusError
			report.Error = err.Error()
			worst = max(worst, exitDoctorProbeFailed)
		case agentsummons.VersionNewer(installed, report.Validated):
			report.Installed = installed
			report.Status = statusDrift
			worst = max(worst, exitDoctorDrift)
		default:
			report.Installed = installed
			report.Status = statusClean
		}
		reports = append(reports, report)
	}

	if asJSON {
		if err := writeJSON(cmd, doctorEnvelope{SchemaVersion: schemaVersion, Harnesses: reports}); err != nil {
			return err
		}
	} else {
		for _, r := range reports {
			switch r.Status {
			case statusNotInstalled:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-12s not installed\n", r.Harness)
			case statusError:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-12s version probe failed: %s\n", r.Harness, r.Error)
			case statusDrift:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-12s installed %s > validated %s — drift candidate; check for an agentsummons update\n", r.Harness, r.Installed, r.Validated)
			default:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-12s installed %s, validated %s — clean\n", r.Harness, r.Installed, r.Validated)
			}
		}
	}
	if worst != 0 {
		return silentExit(cmd, worst)
	}
	return nil
}
