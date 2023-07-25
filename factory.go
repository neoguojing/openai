package openai

type IChat interface {
	Complete(string) (string, error)
}
