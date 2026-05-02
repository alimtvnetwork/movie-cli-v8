// Package doctor provides the diagnostic engine for `movie doctor`.
//
// Surfaces the exact failure modes documented in
// spec/09-app-issues/08-updater-stale-handoff-loop-full-rca.md:
//
//  1. Active PATH binary differs from powershell.json deployPath
//  2. Deploy directory is missing from $PATH entirely
//  3. Stale *-update-* handoff workers left on disk
//  4. Active binary version is older than the freshly built one
//
// Each Check returns a Finding so the report and the --fix path can act on
// them uniformly.
package doctor

import (
	"github.com/alimtvnetwork/movie-cli-v8/apperror"
)

// Severity describes how serious a finding is.
type Severity string

const (
	SeverityOK   Severity = "OK"
	SeverityWarn Severity = "WARN"
	SeverityErr  Severity = "ERR"
)

// Finding is the result of one diagnostic check.
type Finding struct {
	ID        string
	Title     string
	Severity  Severity
	Detail    string
	FixHint   string
	IsFixable bool
}

// Report bundles all findings from a Diagnose run.
// Field order optimized for govet fieldalignment (strings first, slice last).
type Report struct {
	Source    string
	Target    string
	DeployDir string
	Findings  []Finding
	Repo      RepoStatus
}

// Diagnose runs every check and returns the aggregated report.
func Diagnose() (*Report, error) {
	report := &Report{}
	if err := populatePaths(report); err != nil {
		return nil, apperror.Wrap("doctor: cannot resolve paths", err)
	}
	report.Findings = append(report.Findings, checkPathMismatch(report))
	report.Findings = append(report.Findings, checkDeployInPath(report))
	report.Findings = append(report.Findings, checkStaleWorkers(report)...)
	report.Findings = append(report.Findings, checkVersionDrift(report))
	runEnvChecks(report)
	report.Repo = CheckRepo()
	return report, nil
}

// HasFixable returns true if any finding can be auto-repaired.
func (r *Report) HasFixable() bool {
	for _, f := range r.Findings {
		if f.IsFixable && f.Severity != SeverityOK {
			return true
		}
	}
	return false
}

// HasErrors returns true if any finding is at ERR severity.
func (r *Report) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityErr {
			return true
		}
	}
	return false
}
