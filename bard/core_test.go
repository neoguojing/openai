package bard

import (
	"testing"
)

func TestBard(t *testing.T) {
	token := "ZAgNd-O-K-y8Qt4qcvDCNV1hH-qn7xGpDt8Pji_T6-psf39C73zCqCUTuOn3TvTay6pWfg."
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
