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
	InputFile string   `yaml:"input_file"`
	ExpectFile string  `yaml:"expected_file"`
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

func firstDiffLine(got, expected string) (line int, gotLine, expectedLine string) {
	gs := strings.Split(got, "\n")
	es := strings.Split(expected, "\n")
	n := len(gs)
	if len(es) > n {
		n = len(es)
	}
	for i := 0; i < n; i++ {
		var g, e string
		if i < len(gs) {
			g = gs[i]
		}
		if i < len(es) {
			e = es[i]
		}
		if g != e {
			return i + 1, g, e
		}
	}
	return 0, "", ""
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

	h := NewMDHandler()

	for _, tc := range tf.Cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			input := tc.Input
			expected := tc.Expected
			if tc.InputFile != "" {
				b, err := os.ReadFile(filepath.Join("testdata", tc.InputFile))
				if err != nil {
					t.Fatalf("read input_file: %v", err)
				}
				input = string(b)
			}
			if tc.ExpectFile != "" {
				b, err := os.ReadFile(filepath.Join("testdata", tc.ExpectFile))
				if err != nil {
					t.Fatalf("read expected_file: %v", err)
				}
				expected = string(b)
			}

			out := h.Handle(&tc.MDConfig, []byte(input), tc.Direction)
			gotStr := trimRightSpacesPerLine(string(out))
			expectedStr := trimRightSpacesPerLine(expected)
			if gotStr != expectedStr {
				line, gotLine, expLine := firstDiffLine(gotStr, expectedStr)
				if line == 0 {
					t.Fatalf("unexpected output")
				}
				t.Fatalf("unexpected output at line %d\n--- got ---\n%s\n--- expected ---\n%s", line, gotLine, expLine)
			}
		})
	}
}

func TestCountLeadingSpaces(t *testing.T) {
	if got := countLeadingSpaces([]byte("")); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
	if got := countLeadingSpaces([]byte("   x")); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
	if got := countLeadingSpaces([]byte("\tx")); got != 0 {
		t.Fatalf("expected 0 for tab-leading, got %d", got)
	}
}

func TestParseFenceLine(t *testing.T) {
	ch, ln, indent, ok := parseFenceLine([]byte("```go"))
	if !ok || ch != '`' || ln != 3 || indent != 0 {
		t.Fatalf("unexpected parse: ok=%v ch=%q ln=%d indent=%d", ok, ch, ln, indent)
	}

	ch, ln, indent, ok = parseFenceLine([]byte("  ~~~"))
	if !ok || ch != '~' || ln != 3 || indent != 2 {
		t.Fatalf("unexpected parse: ok=%v ch=%q ln=%d indent=%d", ok, ch, ln, indent)
	}

	_, _, _, ok = parseFenceLine([]byte("``"))
	if ok {
		t.Fatalf("expected not ok for short fence")
	}

	_, _, _, ok = parseFenceLine([]byte(" abc"))
	if ok {
		t.Fatalf("expected not ok for non-fence")
	}
}

func TestFenceDeltaForLine(t *testing.T) {
	stack := []mdListStackEntry{{rawIndent: 0, indent: 0, hadContinuation: false, sawMarkerLine: true}}
	if got := fenceDeltaForLine(stack, []byte("```")); got != 4 {
		t.Fatalf("expected delta 4, got %d", got)
	}
	if got := fenceDeltaForLine(stack, []byte("    ```")); got != 0 {
		t.Fatalf("expected delta 0, got %d", got)
	}
	if got := fenceDeltaForLine(nil, []byte("```")); got != 0 {
		t.Fatalf("expected delta 0 with empty stack, got %d", got)
	}
}

func TestShouldClearListStack(t *testing.T) {
	stack := []mdListStackEntry{{rawIndent: 0, indent: 0, hadContinuation: false, sawMarkerLine: true}}
	if got := shouldClearListStack(stack, []byte("Content"), false); !got {
		t.Fatalf("expected to clear list stack")
	}
	if got := shouldClearListStack(stack, []byte("  Content"), false); got {
		t.Fatalf("expected not to clear when indented")
	}
	if got := shouldClearListStack(stack, []byte("- item"), false); got {
		t.Fatalf("expected not to clear for list item start")
	}
	if got := shouldClearListStack(stack, []byte("# Heading"), false); got {
		t.Fatalf("expected not to clear for heading")
	}
}

func TestHandleListItemStart(t *testing.T) {
	var stack []mdListStackEntry
	stack, desired, needBlank := handleListItemStart(stack, 0, false)
	if desired != 0 || needBlank {
		t.Fatalf("expected desired=0 needBlank=false, got desired=%d needBlank=%v", desired, needBlank)
	}
	stack = append(stack, mdListStackEntry{rawIndent: 0, indent: 0, hadContinuation: true, sawMarkerLine: true})
	_, desired, needBlank = handleListItemStart(stack, 0, false)
	if desired != 0 || !needBlank {
		t.Fatalf("expected desired=0 needBlank=true for sibling after continuation, got desired=%d needBlank=%v", desired, needBlank)
	}

	stack = []mdListStackEntry{{rawIndent: 0, indent: 0, hadContinuation: false, sawMarkerLine: true}}
	_, desired, needBlank = handleListItemStart(stack, 2, true)
	if desired != 4 || needBlank {
		t.Fatalf("expected desired=4 needBlank=false for nested, got desired=%d needBlank=%v", desired, needBlank)
	}
}
