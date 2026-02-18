package md

import (
	"bytes"

	log "github.com/sirupsen/logrus"
)

type MDHandler struct {
	handlers []MDSourceHandler
}

func NewMDHandler() *MDHandler {
	return &MDHandler{handlers: []MDSourceHandler{&MDNormalizeHandler{}}}
}

func (m *MDHandler) Handle(mdConfig *MDConfig, mdBytes []byte, direction string) []byte {
	if len(mdBytes) == 0 || mdConfig == nil {
		log.Warnf("md handler: mdBytes is empty or mdConfig is nil")
		return mdBytes
	}

	log.Infof("md handler: handling mdBytes with direction %s", direction)

	frontMatter, body := splitYAMLFrontMatter(mdBytes)

	for _, handler := range m.handlers {
		body = handler.Handle(mdConfig, body, direction)
	}

	if len(frontMatter) == 0 {
		return body
	}
	return append(frontMatter, body...)

}

func splitYAMLFrontMatter(mdBytes []byte) (frontMatter []byte, body []byte) {
	if len(mdBytes) < 4 {
		return nil, mdBytes
	}
	if !bytes.HasPrefix(mdBytes, []byte("---")) {
		return nil, mdBytes
	}
	if len(mdBytes) >= 4 && mdBytes[3] != '\n' && mdBytes[3] != '\r' {
		return nil, mdBytes
	}

	idx := bytes.Index(mdBytes, []byte("\n---"))
	altIdx := bytes.Index(mdBytes, []byte("\n..."))
	closeIdx := -1
	if idx >= 0 {
		closeIdx = idx + 1
	}
	if altIdx >= 0 && (closeIdx == -1 || altIdx+1 < closeIdx) {
		closeIdx = altIdx + 1
	}
	if closeIdx == -1 {
		return nil, mdBytes
	}

	end := bytes.IndexByte(mdBytes[closeIdx:], '\n')
	if end == -1 {
		return mdBytes, nil
	}
	end = closeIdx + end + 1
	return mdBytes[:end], mdBytes[end:]
}
