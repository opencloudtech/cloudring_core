package roadmaprequirements

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var (
	coverageRequirement = regexp.MustCompile(`(?m)^\| (CR-[A-Z0-9]+-[0-9]{3}) \|`)
	definedRequirement  = regexp.MustCompile("(?m)^\\| `?(CR-[A-Z0-9]+-[0-9]{3})`? \\|")
)

func TestRoadmapCoverageReferencesDefinedPublicRequirements(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))

	coverage := readFile(t, filepath.Join(root, "roadmap", "COVERAGE.md"))
	definitions := map[string][]string{}
	goalRequirementsPath := filepath.Join(root, "specifications", "goal-01.md")
	for _, match := range definedRequirement.FindAllStringSubmatch(readFile(t, goalRequirementsPath), -1) {
		definitions[match[1]] = append(definitions[match[1]], filepath.Base(goalRequirementsPath))
	}

	missing := []string{}
	coverageCounts := map[string]int{}
	for _, match := range coverageRequirement.FindAllStringSubmatch(coverage, -1) {
		coverageCounts[match[1]]++
		files := definitions[match[1]]
		if len(files) != 1 {
			missing = append(missing, match[1]+"="+strings.Join(files, ","))
		}
	}
	for id, count := range coverageCounts {
		if count != 1 {
			missing = append(missing, id+"=coverage-count-"+strconv.Itoa(count))
		}
	}
	for id, files := range definitions {
		if len(files) != 1 {
			missing = append(missing, id+"=definition-count-"+strconv.Itoa(len(files)))
		}
		if coverageCounts[id] != 1 {
			missing = append(missing, id+"=uncovered")
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("roadmap coverage requirements must have exactly one public definition: %s", strings.Join(missing, "; "))
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
