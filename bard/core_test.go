package bard

import (
	"os"
	"testing"
)

func TestBard(t *testing.T) {

	token := os.Getenv("_BARD_API_KEY")
	bard := NewBard(token, 30, nil, nil, "", "en", false, "")

	bardAnswer, err := bard.GetAnswer("Hello!")
	if err != nil {
		t.Error(err)
	}

	t.Log(bardAnswer)

	audio, err := bard.Speech("Tell me a joke.", "en-US")
	if err != nil {
		t.Error(err)
	}

	t.Log(audio)
	cookie := bard.ExtractCookie()
	t.Log(cookie)
}
