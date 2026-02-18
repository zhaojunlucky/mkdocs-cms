package md

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type normalizeTestFile struct {
	Cases []normalizeCase `yaml:"cases"`
}

type normalizeCase struct {
	Name      string   `yaml:"name"`
	Direction string   `yaml:"direction"`
	MDConfig  MDConfig `yaml:"md_config"`
	Input     string   `yaml:"input"`
	Expected  string   `yaml:"expected"`
}

func trimRightSpacesPerLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.Join(lines, "\n")
}

func TestMDNormalizeHandler_FromYAMLFixtures(t *testing.T) {
	fixturePath := filepath.Join("testdata", "md_normalize_cases.yml")
	b, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var tf normalizeTestFile
	if err := yaml.Unmarshal(b, &tf); err != nil {
		t.Fatalf("unmarshal fixture yaml: %v", err)
	}
	if len(tf.Cases) == 0 {
		t.Fatalf("no cases in fixture")
	}

	h := &MDNormalizeHandler{}

	for _, tc := range tf.Cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			out := h.Handle(&tc.MDConfig, []byte(tc.Input), tc.Direction)
			gotStr := trimRightSpacesPerLine(string(out))
			expectedStr := trimRightSpacesPerLine(tc.Expected)
			if gotStr != expectedStr {
				t.Fatalf("unexpected output\n--- got ---\n%s\n--- expected ---\n%s", gotStr, expectedStr)
			}
		})
	}
}
