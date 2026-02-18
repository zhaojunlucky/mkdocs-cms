package md

import (
	"bytes"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MDNormalizeHandler struct {
}

type mdListEditSpan struct {
	start, end  int
	replacement []byte
}

type mdListStackEntry struct {
	rawIndent       int
	indent          int
	hadContinuation bool
	sawMarkerLine   bool
}

func (m *MDNormalizeHandler) Handle(mdConfig *MDConfig, mdBytes []byte, direction string) []byte {
	if len(mdBytes) == 0 {
		return mdBytes
	}

	mdBytes = m.normalizeCodeBlocks(mdConfig, mdBytes, direction)
	mdBytes = m.normalizeMkDocsListFormatting(mdBytes, direction)
	return mdBytes
}

func (m *MDNormalizeHandler) normalizeCodeBlocks(mdConfig *MDConfig, mdBytes []byte, direction string) []byte {
	if mdConfig == nil {
		return mdBytes
	}
	if len(mdConfig.CodeBlockTransforms) == 0 || direction == DirectionBoth {
		return mdBytes
	}

	rd := text.NewReader(mdBytes)
	doc := goldmark.DefaultParser().Parse(rd)

	type span struct {
		start, end  int
		replacement string
	}
	var toReplace []span

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		fcb, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}

		lang := string(fcb.Language(mdBytes))
		for _, cb := range mdConfig.CodeBlockTransforms {
			if !cb.IsEnabled() {
				continue
			}
			if !cb.CheckDirection(direction) {
				continue
			}
			fromLang, toLang := cb.GetLang(direction)
			if lang == fromLang {
				toReplace = append(toReplace, span{start: fcb.Info.Segment.Start, end: fcb.Info.Segment.Stop, replacement: toLang})
			}
		}
		return ast.WalkContinue, nil
	})

	if len(toReplace) == 0 {
		return mdBytes
	}

	out := make([]byte, 0, len(mdBytes)+len(toReplace)*5)
	cursor := 0
	for _, sp := range toReplace {
		out = append(out, mdBytes[cursor:sp.start]...)
		out = append(out, []byte(sp.replacement)...)
		cursor = sp.end
	}
	out = append(out, mdBytes[cursor:]...)
	return out
}

func normalizeListSpan(src []byte, spanStart, spanEnd int, edits *[]mdListEditSpan) {
	chunk := src[spanStart:spanEnd]
	lineAbsStart := spanStart

	var stack []mdListStackEntry

	prevLineWasBlank := true
	inFenced := false
	fenceChar := byte(0)
	fenceLen := 0
	fenceDelta := 0

	for len(chunk) > 0 {
		idx := bytes.IndexByte(chunk, '\n')
		var line []byte
		var rest []byte
		lineLen := 0
		if idx == -1 {
			line = chunk
			rest = nil
			lineLen = len(chunk)
		} else {
			line = chunk[:idx]
			rest = chunk[idx+1:]
			lineLen = idx
		}

		lineNoCR := bytes.TrimRight(line, "\r")
		trimSpace := bytes.TrimSpace(lineNoCR)
		isBlank := len(trimSpace) == 0

		if shouldClearListStack(stack, lineNoCR, isBlank) {
			stack = nil
		}

		isFenceLine := false
		isClosingFenceLine := false
		fch, fln, _, fok := parseFenceLine(lineNoCR)
		if fok {
			isFenceLine = true
			if !inFenced {
				inFenced = true
				fenceChar = fch
				fenceLen = fln
				fenceDelta = fenceDeltaForLine(stack, lineNoCR)
			} else if fch == fenceChar && fln >= fenceLen {
				isClosingFenceLine = true
			}
		}

		handled := false

		// Headings should never be treated as list continuation. CommonMark allows up to 3 leading
		// spaces before an ATX heading marker.
		if isATXHeadingLine(lineNoCR) {
			stack = nil
			prevLineWasBlank = isBlank
			handled = true
		}

		if !handled {
			indent, isItemStart := parseListItemIndentBytes(lineNoCR)
			if isItemStart {
				rawIndent := indent
				var desiredIndent int
				needBlankBefore := false
				stack, desiredIndent, needBlankBefore = handleListItemStart(stack, rawIndent, prevLineWasBlank)

				// Rewrite marker indentation if needed, folding blank line insertion into replacement.
				if desiredIndent != rawIndent || needBlankBefore {
					restOfLine := bytes.TrimLeft(line, " ")
					restOfLine = bytes.TrimRight(restOfLine, "\r")
					prefix := []byte(strings.Repeat(" ", desiredIndent))
					if needBlankBefore {
						prefix = append([]byte("\n"), prefix...)
					}
					repl := append(prefix, restOfLine...)
					*edits = append(*edits, mdListEditSpan{start: lineAbsStart, end: lineAbsStart + lineLen, replacement: repl})
					prevLineWasBlank = false
				}

				stack = append(stack, mdListStackEntry{rawIndent: rawIndent, indent: desiredIndent, hadContinuation: false, sawMarkerLine: true})
				prevLineWasBlank = isBlank
				handled = true
			}
		}

		if !handled {
			if len(stack) > 0 {
				top := &stack[len(stack)-1]
				trimLeft := bytes.TrimLeft(lineNoCR, " ")
				// Any non-empty continuation line should be indented to top.indent+4 for MkDocs/Python-Markdown.
				if len(bytes.TrimSpace(trimLeft)) > 0 {
					desiredIndent := top.indent + 4
					var replPrefix []byte
					var payload []byte
					if inFenced {
						leading := countLeadingSpaces(lineNoCR)
						newLeading := leading + fenceDelta
						if newLeading < 0 {
							newLeading = 0
						}
						replPrefix = []byte(strings.Repeat(" ", newLeading))
						payload = lineNoCR[leading:]
					} else {
						replPrefix = []byte(strings.Repeat(" ", desiredIndent))
						payload = trimLeft
					}
					// Ensure a blank line between the marker line and the first continuation line.
					if top.sawMarkerLine && !top.hadContinuation && !prevLineWasBlank {
						replPrefix = append([]byte("\n"), replPrefix...)
					}
					repl := append(replPrefix, payload...)
					*edits = append(*edits, mdListEditSpan{start: lineAbsStart, end: lineAbsStart + lineLen, replacement: repl})
					top.hadContinuation = true
					prevLineWasBlank = false
					handled = true
				}
			}
		}

		if inFenced && isFenceLine && isClosingFenceLine {
			inFenced = false
			fenceChar = 0
			fenceLen = 0
			fenceDelta = 0
		}

		if !handled {
			prevLineWasBlank = isBlank
		}

		if idx == -1 {
			break
		}
		lineAbsStart = lineAbsStart + lineLen + 1
		chunk = rest
	}
}

func countLeadingSpaces(b []byte) int {
	i := 0
	for i < len(b) && b[i] == ' ' {
		i++
	}
	return i
}

func parseFenceLine(b []byte) (fenceChar byte, fenceLen int, indent int, ok bool) {
	indent = countLeadingSpaces(b)
	if indent >= len(b) {
		return 0, 0, 0, false
	}
	ch := b[indent]
	if ch != '`' && ch != '~' {
		return 0, 0, 0, false
	}
	k := indent
	for k < len(b) && b[k] == ch {
		k++
	}
	if k-indent < 3 {
		return 0, 0, 0, false
	}
	return ch, k - indent, indent, true
}

func fenceDeltaForLine(stack []mdListStackEntry, lineNoCR []byte) int {
	if len(stack) == 0 {
		return 0
	}
	desiredIndent := stack[len(stack)-1].indent + 4
	existingIndent := countLeadingSpaces(lineNoCR)
	return desiredIndent - existingIndent
}

func shouldClearListStack(stack []mdListStackEntry, lineNoCR []byte, isBlank bool) bool {
	if len(stack) == 0 || isBlank {
		return false
	}
	leading := countLeadingSpaces(lineNoCR)
	if leading > stack[0].rawIndent {
		return false
	}
	if _, isItemStart := parseListItemIndentBytes(lineNoCR); isItemStart {
		return false
	}
	if isATXHeadingLine(lineNoCR) {
		return false
	}
	return true
}

func handleListItemStart(stack []mdListStackEntry, rawIndent int, prevLineWasBlank bool) (newStack []mdListStackEntry, desiredIndent int, needBlankBefore bool) {
	needBlankBefore = false
	desiredIndent = normalizeIndentTo4(rawIndent)

	if len(stack) > 0 {
		top := &stack[len(stack)-1]
		if rawIndent > top.rawIndent {
			top.hadContinuation = true
			desiredIndent = top.indent + 4
		} else {
			prevHadCont := false
			for len(stack) > 0 && rawIndent <= stack[len(stack)-1].rawIndent {
				prevHadCont = stack[len(stack)-1].hadContinuation
				stack = stack[:len(stack)-1]
			}
			if prevHadCont {
				needBlankBefore = !prevLineWasBlank
			}
			if len(stack) > 0 && desiredIndent < stack[len(stack)-1].indent+4 {
				desiredIndent = stack[len(stack)-1].indent + 4
			}
		}
	}

	return stack, desiredIndent, needBlankBefore
}

func parseListItemIndentBytes(line []byte) (int, bool) {
	i := 0
	for i < len(line) && line[i] == ' ' {
		i++
	}
	if i >= len(line) {
		return 0, false
	}
	if (line[i] == '-' || line[i] == '*' || line[i] == '+') && i+1 < len(line) && line[i+1] == ' ' {
		return i, true
	}
	j := i
	for j < len(line) && line[j] >= '0' && line[j] <= '9' {
		j++
	}
	if j == i {
		return 0, false
	}
	if j+1 < len(line) && line[j] == '.' && line[j+1] == ' ' {
		return i, true
	}
	return 0, false
}

func normalizeIndentTo4(indent int) int {
	if indent <= 0 {
		return 0
	}
	// Python-Markdown is strict about 4-space indentation for list continuation and nested lists.
	// Round up to the next multiple of 4.
	return ((indent + 3) / 4) * 4
}

func isATXHeadingLine(line []byte) bool {
	i := 0
	for i < len(line) && line[i] == ' ' {
		i++
	}
	if i >= len(line) {
		return false
	}

	start := i
	for i < len(line) && i-start < 6 && line[i] == '#' {
		i++
	}
	if i == start {
		return false
	}
	// Too many leading '#'
	if i-start > 6 {
		return false
	}
	// Require at least one whitespace after the 1..6 '#'
	if i >= len(line) {
		return false
	}
	return line[i] == ' ' || line[i] == '\t'
}

func (m *MDNormalizeHandler) normalizeMkDocsListFormatting(mdBytes []byte, direction string) []byte {
	if len(mdBytes) == 0 || direction != DirectionWrite {
		return mdBytes
	}

	rd := text.NewReader(mdBytes)
	doc := goldmark.DefaultParser().Parse(rd)

	var hasList bool
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if _, ok := n.(*ast.List); ok {
				hasList = true
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	})
	if !hasList {
		return mdBytes
	}

	var edits []mdListEditSpan
	normalizeListSpan(mdBytes, 0, len(mdBytes), &edits)

	if len(edits) == 0 {
		return mdBytes
	}

	sort.Slice(edits, func(i, j int) bool {
		if edits[i].start == edits[j].start {
			return edits[i].end > edits[j].end
		}
		return edits[i].start < edits[j].start
	})

	out := make([]byte, 0, len(mdBytes)+len(edits)*4)
	cursor := 0
	for _, e := range edits {
		if e.start < cursor {
			continue
		}
		out = append(out, mdBytes[cursor:e.start]...)
		out = append(out, e.replacement...)
		cursor = e.end
	}
	out = append(out, mdBytes[cursor:]...)
	return out
}
