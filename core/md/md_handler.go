package md

import log "github.com/sirupsen/logrus"

type MDHandler struct {
	mdConfig *MDConfig
	handlers []MDSourceHandler
}

func NewMDHandler(mdConfig *MDConfig) *MDHandler {
	return &MDHandler{mdConfig: mdConfig, handlers: []MDSourceHandler{NewMDCodeBlockHandler(mdConfig)}}
}

func (m *MDHandler) Handle(mdBytes []byte, direction string) []byte {
	if len(mdBytes) == 0 || m.mdConfig == nil {
		log.Warnf("md handler: mdBytes is empty or mdConfig is nil")
		return mdBytes
	}

	log.Infof("md handler: handling mdBytes with direction %s", direction)

	for _, handler := range m.handlers {
		mdBytes = handler.Handle(mdBytes, direction)
	}

	return mdBytes

}
