package md

import (
	log "github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type MDCodeBlockHandler struct {
	mdConfig *MDConfig
}

func NewMDCodeBlockHandler(mdConfig *MDConfig) *MDCodeBlockHandler {
	return &MDCodeBlockHandler{mdConfig: mdConfig}
}

func (m *MDCodeBlockHandler) Handle(mdBytes []byte, direction string) []byte {
	if len(m.mdConfig.CodeBlockTransforms) == 0 {
		return mdBytes
	}

	rd := text.NewReader(mdBytes)
	doc := goldmark.DefaultParser().Parse(rd)

	// We'll collect replacements as (start,end) byte ranges for the "info" segment.
	type span struct {
		start, end int
		cfg        CodeBlockTransform
	} // inclusive start, exclusive end
	var toReplace []span

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if fcb, ok := n.(*ast.FencedCodeBlock); ok {
			lang := string(fcb.Language(mdBytes)) // content of the "info" token
			for _, cb := range m.mdConfig.CodeBlockTransforms {
				if !cb.CheckDirection(direction) {
					continue
				}
				if lang == cb.FromLang {
					toReplace = append(toReplace, span{start: fcb.Info.Segment.Start, end: fcb.Info.Segment.Stop, cfg: cb})
				}
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		log.Errorf("failed to walk md: %v", err)
	}

	if len(toReplace) == 0 {
		return mdBytes // nothing to do
	}

	// Build a new buffer with replacements applied.
	out := make([]byte, 0, len(mdBytes)+len(toReplace)*5)
	cursor := 0
	for _, sp := range toReplace {
		// Copy everything before the info token
		out = append(out, mdBytes[cursor:sp.start]...)
		// Insert the new token
		out = append(out, []byte("kroki-mermaid")...)
		// Skip the old token
		cursor = sp.end
	}
	// Copy the remainder
	out = append(out, mdBytes[cursor:]...)
	return out

}
