package md

import "testing"

func TestMDHandler_SkipYAMLFrontMatter(t *testing.T) {
	mdConfig := &MDConfig{CodeBlockTransforms: []CodeBlockTransform{}}
	h := NewMDHandler()

	in := "---\n" +
		"title: x\n" +
		"list: [1, 2]\n" +
		"---\n" +
		"1. p\n" +
		"   - c1\n" +
		"     text\n"

	out := string(h.Handle(mdConfig, []byte(in), DirectionWrite))

	expected := "---\n" +
		"title: x\n" +
		"list: [1, 2]\n" +
		"---\n" +
		"1. p\n" +
		"    - c1\n" +
		"\n" +
		"        text\n"

	if out != expected {
		t.Fatalf("unexpected output\n--- got ---\n%s\n--- expected ---\n%s", out, expected)
	}
}
