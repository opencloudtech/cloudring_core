package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/opencloudtech/CloudRING/pkg/ocs"
	"github.com/opencloudtech/CloudRING/sdk/ocsv3"
)

func main() {
	if err := runWithIO(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithIO(args, os.Stdout, os.Stderr)
}

func runWithIO(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		return usageError()
	}
	switch args[0] {
	case "validate":
		return runValidate(args[1:], stdout)
	case "conformance":
		return runConformance(args[1:], stdout, stderr)
	default:
		return usageError()
	}
}

func runValidate(paths []string, stdout io.Writer) error {
	if len(paths) == 0 {
		return usageError()
	}

	var failed bool
	for _, path := range paths {
		input := openOperatorSelectedInput(path)
		subject, validationErr := validateOperatorSelectedInput(input)
		closeErr := input.close()
		if validationErr != nil {
			fmt.Fprintf(stdout, "ocs_connector_package_invalid %s: %v\n", subject, validationErr)
			failed = true
			if closeErr != nil {
				failed = true
			}
			continue
		}
		if closeErr != nil {
			fmt.Fprintf(stdout, "ocs_connector_package_invalid %s: selected connector package could not be closed safely\n", subject)
			failed = true
			continue
		}
		fmt.Fprintf(stdout, "ocs_connector_package_valid %s\n", subject)
	}

	if failed {
		return errors.New("one or more connector packages are invalid")
	}
	return nil
}

func runConformance(args []string, stdout io.Writer, stderr io.Writer) (resultErr error) {
	paths, evidencePath, err := parseConformanceArgs(args)
	if err != nil {
		return err
	}
	inputs := make([]*operatorSelectedInput, 0, len(paths))
	for _, path := range paths {
		inputs = append(inputs, openOperatorSelectedInput(path))
	}
	defer func() {
		for _, input := range inputs {
			if err := input.close(); err != nil {
				resultErr = errors.Join(resultErr, errors.New("close selected connector package safely"))
			}
		}
	}()

	var guard *evidenceInputGuard
	if evidencePath != "" {
		guard, err = newEvidenceInputGuard(evidencePath, inputs)
		if err != nil {
			return err
		}
	}
	reports := make([]ocsv3.ConformanceReport, 0, len(paths))
	var failed bool
	for _, input := range inputs {
		report, err := conformanceReportForInput(input)
		if err != nil {
			fmt.Fprintf(stdout, "ocs_conformance_failed %s: %v\n", input.opaqueID, err)
			reports = append(reports, ocsv3.ConformanceReport{PackageName: input.opaqueID, Passed: false, Summary: err.Error()})
			failed = true
			continue
		}
		reports = append(reports, report)
		if report.Passed {
			fmt.Fprintf(stdout, "ocs_conformance_passed %s: %s\n", report.PackageName, report.Summary)
			continue
		}
		fmt.Fprintf(stdout, "ocs_conformance_failed %s: %s\n", report.PackageName, report.Summary)
		for _, problem := range report.Problems {
			fmt.Fprintf(stdout, "- %s %s: %s remediation=%s\n", problem.Surface, problem.Field, problem.Message, problem.Remediation)
		}
		failed = true
	}
	if evidencePath != "" {
		if err := writeConformanceEvidenceWithGuard(evidencePath, reports, guard); err != nil {
			return fmt.Errorf("write conformance evidence: %w", err)
		}
	}
	if !failed {
		return nil
	}
	return fmt.Errorf("conformance failed for %d connector package(s)", failedConformanceReports(reports))
}

func conformanceReportForInput(input *operatorSelectedInput) (ocsv3.ConformanceReport, error) {
	if input == nil || input.readErr != nil {
		return ocsv3.ConformanceReport{}, errors.New("read connector package: selected input is unavailable")
	}
	pkg, err := ocsv3.ParseConnectorPackage(input.data)
	if err != nil {
		return ocsv3.ConformanceReport{}, err
	}
	report := ocsv3.CheckConformance(pkg)
	report.PackageName = input.opaqueID
	if report.Passed {
		report.Summary = "OCSv3 conformance passed for " + input.opaqueID
	} else {
		report.Summary = fmt.Sprintf("OCSv3 conformance failed for %s with %d problem(s)", input.opaqueID, len(report.Problems))
	}
	sanitizeConformanceReport(&report)
	return report, nil
}

func parseConformanceArgs(args []string) ([]string, string, error) {
	paths := make([]string, 0)
	var evidencePath string
	operandsOnly := false
	remaining := args
	for len(remaining) > 0 {
		arg := remaining[0]
		remaining = remaining[1:]
		if operandsOnly {
			paths = append(paths, arg)
			continue
		}
		if arg == "--" {
			operandsOnly = true
			continue
		}
		if arg == "--evidence" {
			if len(remaining) == 0 {
				return nil, "", errors.New("usage: --evidence requires a path")
			}
			evidencePath = remaining[0]
			remaining = remaining[1:]
			if evidencePath == "" {
				return nil, "", errors.New("usage: --evidence requires a path")
			}
			continue
		}
		if strings.HasPrefix(arg, "--evidence=") {
			evidencePath = strings.TrimPrefix(arg, "--evidence=")
			if evidencePath == "" {
				return nil, "", errors.New("usage: --evidence requires a path")
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return nil, "", errors.New("unknown conformance flag")
		}
		paths = append(paths, arg)
	}
	if len(paths) == 0 {
		return nil, "", errors.New("usage: ocsctl conformance <connector-package.json> [connector-package.json...] [--evidence report.json]")
	}
	return paths, evidencePath, nil
}

func failedConformanceReports(reports []ocsv3.ConformanceReport) int {
	var failed int
	for _, report := range reports {
		if !report.Passed {
			failed++
		}
	}
	return failed
}

func usageError() error {
	return errors.New("usage: ocsctl validate <connector-package.json> [connector-package.json...] OR ocsctl conformance <connector-package.json> [--evidence report.json]")
}

func writeConformanceEvidence(path string, reports []ocsv3.ConformanceReport) error {
	return writeConformanceEvidenceWithGuard(path, reports, nil)
}

func writeConformanceEvidenceWithGuard(path string, reports []ocsv3.ConformanceReport, guard *evidenceInputGuard) error {
	if guard != nil {
		if err := guard.verify(); err != nil {
			return err
		}
	}
	if err := ensureEvidenceParentDirectory(path); err != nil {
		return err
	}
	var (
		data []byte
		err  error
	)
	if len(reports) == 1 {
		data, err = json.MarshalIndent(reports[0], "", "  ")
	} else {
		data, err = json.MarshalIndent(reports, "", "  ")
	}
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if guard == nil {
		return writePrivateFileSafely(path, data)
	}
	hooks := evidenceWriteHooks{
		beforeReplaceValidation: func(_, _ string) error { return guard.verify() },
		beforePublish:           func(_, _ string) error { return guard.verify() },
	}
	return writePrivateFileSafelyWithHooks(path, data, replaceEvidenceFile, hooks)
}

func validateOperatorSelectedInput(input *operatorSelectedInput) (string, error) {
	if input == nil {
		return opaqueInputIdentity("unavailable", nil), errors.New("selected connector package is unavailable")
	}
	subject := input.opaqueID
	if input.readErr != nil {
		return subject, input.readErr
	}

	var pkg ocs.ConnectorPackage
	if err := json.Unmarshal(input.data, &pkg); err != nil {
		return subject, fmt.Errorf("parse connector package: %w", err)
	}
	return subject, pkg.Validate()
}

func sanitizeConformanceReport(report *ocsv3.ConformanceReport) {
	if report == nil {
		return
	}
	if !isSafeDeclaredPackageVersion(report.PackageVersion) {
		report.PackageVersion = "redacted"
	}
	for index := range report.Problems {
		originalField := report.Problems[index].Field
		sanitizedField := sanitizeConformanceProblemField(originalField)
		if sanitizedField == originalField {
			continue
		}
		report.Problems[index].Field = sanitizedField
		report.Problems[index].Message = strings.ReplaceAll(report.Problems[index].Message, originalField, sanitizedField)
		report.Problems[index].Remediation = strings.ReplaceAll(report.Problems[index].Remediation, originalField, sanitizedField)
	}
}

func sanitizeConformanceProblemField(field string) string {
	const lifecyclePrefix = "service.spec.lifecycle."
	if strings.HasPrefix(field, lifecyclePrefix) {
		for _, suffix := range []string{".idempotencyKey", ".idempotent"} {
			if !strings.HasSuffix(field, suffix) {
				continue
			}
			name := strings.TrimSuffix(strings.TrimPrefix(field, lifecyclePrefix), suffix)
			if name != "" && !isKnownLifecycleName(name) {
				return lifecyclePrefix + "custom-action" + suffix
			}
			return field
		}
	}
	const statePrefix = "service.spec.states."
	if strings.HasPrefix(field, statePrefix) {
		for _, suffix := range []string{".evidenceRef", ".remediation"} {
			if !strings.HasSuffix(field, suffix) {
				continue
			}
			name := strings.TrimSuffix(strings.TrimPrefix(field, statePrefix), suffix)
			if name != "" && !isKnownStateName(name) {
				return statePrefix + "custom-state" + suffix
			}
			return field
		}
	}
	return field
}

func isKnownLifecycleName(name string) bool {
	switch name {
	case "provision", "backup", "restore", "export", "delete", "retry", "rollback", "repair":
		return true
	default:
		return false
	}
}

func isKnownStateName(name string) bool {
	switch name {
	case "ready", "denied", "degraded", "blocked", "retryable":
		return true
	default:
		return false
	}
}

func isSafeDeclaredPackageVersion(version string) bool {
	if version == "" || len(version) > 64 || strings.TrimSpace(version) != version {
		return false
	}
	core := strings.TrimPrefix(version, "v")
	if core == version && strings.HasPrefix(version, "V") {
		return false
	}
	prerelease := ""
	if separator := strings.IndexByte(core, '-'); separator >= 0 {
		prerelease = core[separator+1:]
		if prerelease == "" {
			return false
		}
		core = core[:separator]
	}
	components := strings.Split(core, ".")
	if len(components) != 3 {
		return false
	}
	for _, component := range components {
		if component == "" || len(component) > 9 {
			return false
		}
		for _, character := range component {
			if character < '0' || character > '9' {
				return false
			}
		}
	}
	if prerelease == "" {
		return true
	}
	parts := strings.Split(prerelease, ".")
	if len(parts) > 2 || (parts[0] != "alpha" && parts[0] != "beta" && parts[0] != "rc") {
		return false
	}
	if len(parts) == 1 {
		return true
	}
	if parts[1] == "" || len(parts[1]) > 9 {
		return false
	}
	for _, character := range parts[1] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
