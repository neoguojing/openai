package bard

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BardCookies struct {
	CookieDict      map[string]string
	Timeout         int
	Proxies         map[string]string
	Session         *http.Client
	Language        string
	RunCode         bool
	Reqid           int
	ConversationId  string
	ResponseId      string
	ChoiceId        string
	SNlM0e          string
}

func NewBardCookies(cookieDict map[string]string, timeout int, proxies map[string]string, session *http.Client, language string, runCode bool) *BardCookies {
	reqid := rand.Intn(9999)
	if session == nil {
		session = &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	}
	for k, v := range cookieDict {
		session.Jar.SetCookies(&url.URL{Scheme: "https", Host: "bard.google.com"}, []*http.Cookie{
			{Name: k, Value: v},
		})
	}
	sNlM0e, _ := getSNlM0e(session, timeout, proxies)
	if language == "" {
		language = os.Getenv("_BARD_API_LANG")
	}
	return &BardCookies{
		CookieDict:     cookieDict,
		Timeout:        timeout,
		Proxies:        proxies,
		Session:        session,
		Language:       language,
		RunCode:        runCode,
		Reqid:          reqid,
		ConversationId: "",
		ResponseId:     "",
		ChoiceId:       "",
		SNlM0e:         sNlM0e,
	}
}

func (b *BardCookies) GetAnswer(inputText string) (map[string]interface{}, error) {
	// implementation of GetAnswer function
	params := map[string]string{
		"bl":    "boq_assistant-bard-web-server_20230419.00_p1",
		"_reqid": strconv.Itoa(b.Reqid),
		"rt":    "c",
	}

	// Set language (optional)
	if b.Language != "" && !stringInSlice(b.Language, ALLOWED_LANGUAGES) {
		translatorToEng := NewGoogleTranslator("auto", "en")
		inputText = translatorToEng.Translate(inputText)
	}

	// Make post data structure and insert prompt
	inputTextStruct := [][]interface{}{
		{inputText},
		nil,
		{b.ConversationId, b.ResponseId, b.ChoiceId},
	}
	data := map[string]string{
		"f.req": json.Marshal([]interface{}{nil, json.Marshal(inputTextStruct)}),
		"at":    b.SNlM0e,
	}

	// Get response
	resp, err := b.Session.Post(
		"https://bard.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate",
		params,
		data,
		b.Timeout,
		b.Proxies,
	)

	// Post-processing of response
	respContent, _ := ioutil.ReadAll(resp.Body)
	respDict := json.Unmarshal(strings.Split(string(respContent), "\n")[3])[0][2]

	if respDict == nil {
		return map[string]interface{}{"content": fmt.Sprintf("Response Error: %s.", respContent)}
	}
	respJson := json.Unmarshal(respDict)

	// Gather image links
	images := make(map[string]struct{})
	if len(respJson) >= 3 {
		if len(respJson[4][0]) >= 4 && respJson[4][0][4] != nil {
			for _, img := range respJson[4][0][4] {
				images[img[0][0][0]] = struct{}{}
			}
		}
	}
	parsedAnswer := json.Unmarshal(respDict)

	// Translated by Google Translator (optional)
	if b.Language != "" && !stringInSlice(b.Language, ALLOWED_LANGUAGES) {
		translatorToLang := NewGoogleTranslator("auto", b.Language)
		parsedAnswer[0][0] = translatorToLang.Translate(parsedAnswer[0][0])
		for i, x := range parsedAnswer[4] {
			parsedAnswer[4][i] = []interface{}{x[0], translatorToLang.Translate(x[1][0])}
		}
	}

	// Get code
	code := ""
	if strings.Contains(parsedAnswer[0][0], "```") {
		code = strings.Split(parsedAnswer[0][0], "```")[1][6:]
	}

	// Return dictionary object
	bardAnswer := map[string]interface{}{
		"content":           parsedAnswer[0][0],
		"conversation_id":   parsedAnswer[1][0],
		"response_id":       parsedAnswer[1][1],
		"factualityQueries": parsedAnswer[3],
		"textQuery":         parsedAnswer[2][0] if parsedAnswer[2] else "",
		"choices":           [{"id": x[0], "content": x[1]} for x in parsedAnswer[4]],
		"links":             b.extractLinks(parsedAnswer[4]),
		"images":            images,
		"code":              code,
	}
	b.ConversationId, b.ResponseId, b.ChoiceId = (
		bardAnswer["conversation_id"],
		bardAnswer["response_id"],
		bardAnswer["choices"][0]["id"],
	)
	b.Reqid += 100000

	// Execute Code
	if b.RunCode && bardAnswer["code"] != "" {
		fmt.Println(bardAnswer["code"])
		// exec(bardAnswer["code"]) // Executing arbitrary code can be dangerous
	}

	return bardAnswer
}

func (b *BardCookies) getSNlM0e() (string, error) {
	resp, err := b.Session.Get("https://bard.google.com/")
	if err != nil || resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Response code not 200. Response Status is %d", resp.StatusCode))
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	sNlM0e := regexp.MustCompile(`SNlM0e\":\"(.*?)\"`).FindStringSubmatch(string(body))
	if len(sNlM0e) < 2 {
		return "", errors.New("SNlM0e value not found in response. Check __Secure-1PSID value.")
	}
	return sNlM0e[1], nil
}

func (b *BardCookies) extractLinks(data []interface{}) []string {
	// implementation of extractLinks function
	var links []string
	if data, ok := data.([]interface{}); ok {
		for _, item := range data {
			if item, ok := item.([]interface{}); ok {
				links = append(links, b.extractLinks(item)...)
			} else if item, ok := item.(string); ok && strings.HasPrefix(item, "http") && !strings.Contains(item, "favicon") {
				links = append(links, item)
			}
		}
	}
	return links
}
