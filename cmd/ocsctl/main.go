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
		if err := validateFile(path); err != nil {
			fmt.Fprintf(stdout, "ocs_connector_package_invalid %s: %v\n", path, err)
			failed = true
			continue
		}
		fmt.Fprintf(stdout, "ocs_connector_package_valid %s\n", path)
	}

	if failed {
		return errors.New("one or more connector packages are invalid")
	}
	return nil
}

func runConformance(args []string, stdout io.Writer, stderr io.Writer) error {
	paths, evidencePath, err := parseConformanceArgs(args)
	if err != nil {
		return err
	}
	reports := make([]ocsv3.ConformanceReport, 0, len(paths))
	var failed bool
	for _, path := range paths {
		report, err := conformanceReportForPath(path)
		if err != nil {
			fmt.Fprintf(stdout, "ocs_conformance_failed %s: %v\n", path, err)
			reports = append(reports, ocsv3.ConformanceReport{PackageName: path, Passed: false, Summary: err.Error()})
			failed = true
			continue
		}
		reports = append(reports, report)
		if report.Passed {
			fmt.Fprintf(stdout, "ocs_conformance_passed %s: %s\n", path, report.Summary)
			continue
		}
		fmt.Fprintf(stdout, "ocs_conformance_failed %s: %s\n", path, report.Summary)
		for _, problem := range report.Problems {
			fmt.Fprintf(stdout, "- %s %s: %s remediation=%s\n", problem.Surface, problem.Field, problem.Message, problem.Remediation)
		}
		failed = true
	}
	if evidencePath != "" {
		if err := writeConformanceEvidence(evidencePath, reports); err != nil {
			return fmt.Errorf("write conformance evidence: %w", err)
		}
	}
	if !failed {
		return nil
	}
	return fmt.Errorf("conformance failed for %d connector package(s)", failedConformanceReports(reports))
}

func conformanceReportForPath(path string) (ocsv3.ConformanceReport, error) {
	data, err := readOperatorSelectedFile(path)
	if err != nil {
		return ocsv3.ConformanceReport{}, fmt.Errorf("read connector package: %w", err)
	}
	pkg, err := ocsv3.ParseConnectorPackage(data)
	if err != nil {
		return ocsv3.ConformanceReport{}, err
	}
	return ocsv3.CheckConformance(pkg), nil
}

func parseConformanceArgs(args []string) ([]string, string, error) {
	paths := make([]string, 0)
	var evidencePath string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--evidence" {
			if i+1 >= len(args) {
				return nil, "", errors.New("usage: --evidence requires a path")
			}
			evidencePath = args[i+1]
			if evidencePath == "" {
				return nil, "", errors.New("usage: --evidence requires a path")
			}
			i++
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
			return nil, "", fmt.Errorf("unknown conformance flag %q", arg)
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
	return writePrivateFileSafely(path, data)
}

func validateFile(path string) error {
	data, err := readOperatorSelectedFile(path)
	if err != nil {
		return err
	}

	var pkg ocs.ConnectorPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}
	return pkg.Validate()
}
