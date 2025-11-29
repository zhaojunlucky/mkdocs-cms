package md

import log "github.com/sirupsen/logrus"

type MDHandler struct {
	handlers []MDSourceHandler
}

func NewMDHandler() *MDHandler {
	return &MDHandler{handlers: []MDSourceHandler{&MDCodeBlockHandler{}}}
}

func (m *MDHandler) Handle(mdConfig *MDConfig, mdBytes []byte, direction string) []byte {
	if len(mdBytes) == 0 || mdConfig == nil {
		log.Warnf("md handler: mdBytes is empty or mdConfig is nil")
		return mdBytes
	}

	log.Infof("md handler: handling mdBytes with direction %s", direction)

	for _, handler := range m.handlers {
		mdBytes = handler.Handle(mdConfig, mdBytes, direction)
	}

	return mdBytes

}
