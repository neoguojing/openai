package main

import (
	"testing"

	"github.com/neoguojing/openai/models"
)

func TestGenerateRecommendationMessage(t *testing.T) {
	userInfo := models.TelegramUserInfo{
		Username: "hello",
		Bio:      "wordl",
	}

	full := &UserInfoFull{
		User: &userInfo,
	}
	resp, err := generateRecommendationMessage(full)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Log(resp)

}
