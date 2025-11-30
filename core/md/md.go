package md

type MDSourceHandler interface {
	Handle(mdCondif *MDConfig, mdBytes []byte, direction string) []byte
}
