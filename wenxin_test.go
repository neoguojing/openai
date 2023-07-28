package openai

import "testing"

func TestBaidu(t *testing.T) {
	c := NewBaiduClient("", "")
	resp, err := c.Complete("介绍下自己")
	if err != nil {
		t.Error(err)
	}

	t.Log(resp)
}
