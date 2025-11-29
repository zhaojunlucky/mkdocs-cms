package md

type MDSourceHandler interface {
	Handle(mdBytes []byte, direction string) []byte
}
