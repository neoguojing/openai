package bard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type BardAsync struct {
	token                                string
	timeout                              time.Duration
	proxies                              map[string]string
	googleTranslatorAPIKey               string
	language                             string
	runCode                              bool
	reqid                                int
	conversationID, responseID, choiceID string
	client                               *http.Client
	SNlM0e                               string
}

func NewBardAsync(token string, timeout int, proxies map[string]string, googleTranslatorAPIKey string, language string, runCode bool) *BardAsync {
	bard := &BardAsync{
		token:                  token,
		timeout:                time.Duration(timeout) * time.Second,
		proxies:                proxies,
		googleTranslatorAPIKey: googleTranslatorAPIKey,
		language:               language,
		runCode:                runCode,
		reqid:                  1000, // This is a placeholder. Replace with your logic.
		client:                 &http.Client{Timeout: time.Duration(timeout) * time.Second},
		SNlM0e:                 "", // This is a placeholder. Replace with your logic.
	}
	return bard
}

func (b *BardAsync) GetAnswer(inputText string) (map[string]interface{}, error) {

	params := map[string]string{
		"bl":     "boq_assistant-bard-web-server_20230419.00_p1",
		"_reqid": fmt.Sprint(b.reqid),
		"rt":     "c",
	}

	// Translation logic goes here

	// Prepare request data
	inputTextStruct := [][]interface{}{{inputText}, nil, {b.conversationID, b.responseID, b.choiceID}}
	data := map[string]interface{}{
		"f.req": inputTextStruct,
		"at":    b.SNlM0e,
	}

	// POST request
	url := "https://bard.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate"
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "__Secure-1PSID", Value: b.token})

	// Send request
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse JSON response
	body, _ := ioutil.ReadAll(resp.Body)
	var respDict map[string]interface{}
	json.Unmarshal(body, &respDict)

	// Further processing and error handling go here

	return respDict, nil
}

// This is a placeholder. Replace with your logic.
// GetSNIM0E is used to get the SNlM0e value from the Bard website.
// The function uses a regular expression to search for the SNlM0e value in the response text.
// If it finds it, then it returns that value.
func (b *Bard) GetSNIM0E() (string, error) {
	if b.Token == "" || b.Token[len(b.Token)-1] != '.' {
		return "", errors.New("__Secure-1PSID value must end with a single dot. Enter correct __Secure-1PSID value.")
	}

	resp, err := b.Client.Get("https://bard.google.com/", b.Timeout, true)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Response code not 200. Response Status is %d", resp.StatusCode)
	}
	snim0e := regexp.MustCompile(`SNlM0e\":\"(.*?)\"`).FindStringSubmatch(resp.Text)
	if snim0e == nil {
		return "", errors.New("SNlM0e value not found in response. Check __Secure-1PSID value.")
	}
	return snim0e[1], nil
}

// ExtractLinks extracts links from the given data.
func (b *Bard) ExtractLinks(data []interface{}) []string {
	links := []string{}
	for _, item := range data {
		switch v := item.(type) {
		case []interface{}:
			links = append(links, b.ExtractLinks(v)...)
		case string:
			if strings.HasPrefix(v, "http") && !strings.Contains(v, "favicon") {
				links = append(links, v)
			}
		}
	}
	return links

}
