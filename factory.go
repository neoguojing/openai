package openai

type IChat interface {
	Complete(string) (*ChatResponse, error)
}
