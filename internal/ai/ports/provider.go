package ports

import "context"

type Request struct{ System, User, Model string }
type Response struct {
	Text                      string
	InputTokens, OutputTokens int
}
type Provider interface {
	Generate(context.Context, Request) (Response, error)
	Name() string
}
type Embedder interface {
	Embed(context.Context, []string, string) ([][]float32, int, error)
}
