package bard

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

type Bard struct {
	Token                  string
	Timeout                int
	Proxies                map[string]string
	Session                *http.Client
	ReqId                  int
	ConversationId         string
	Language               string
	RunCode                bool
	GoogleTranslatorAPIKey string
	SNlM0e                 string
}

func NewBard(token string, timeout int, proxies map[string]string, session *http.Client, conversationId string, language string,
	runCode bool, googleTranslatorAPIKey string) *Bard {
	b := &Bard{
		Token:                  token,
		Timeout:                timeout,
		Proxies:                proxies,
		Session:                session,
		ReqId:                  int(rand.Intn(10000)),
		ConversationId:         conversationId,
		Language:               language,
		RunCode:                runCode,
		GoogleTranslatorAPIKey: googleTranslatorAPIKey,
	}

	if b.Token == "" && tokenFromBrowser {
		b.Token = b.ExtractCookie()
		if b.Token == "" {
			panic("\nCan't extract cookie from browsers.\nPlease sign in first at\nhttps://accounts.google.com/v3/signin/identifier?followup=https://bard.google.com/&flowName=GlifWebSignIn&flowEntry=ServiceLogin")
		}
	}

	if b.Proxies == nil {
		b.Proxies = make(map[string]string)
	}

	if b.Session == nil {
		b.Session = &http.Client{}
		b.Session.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
		b.Session.Jar, _ = cookiejar.New(nil)
		cookie := &http.Cookie{
			Name:  "__Secure-1PSID",
			Value: b.Token,
		}
		b.Session.Jar.SetCookies(url, []*http.Cookie{cookie})
	}

	b.SNlM0e = b.GetSNlM0e()

	return b
}

func (b *Bard) GetAnswer(inputText string) (map[string]interface{},error) {
	// Make POST request and parse response
	// ...
	if b.GoogleTranslatorAPIKey != "" {
		googleOfficialTranslator, err := translate.NewClient(context.Background(), option.WithAPIKey(b.GoogleTranslatorAPIKey))
		if err != nil {
			return nil,errors.New(fmt.Sprintf("Failed to create Google Translator client: %v", err))
		}
	}

	// Set language (optional)
	if b.Language != "" && !contains(ALLOWED_LANGUAGES, b.Language) && b.GoogleTranslatorAPIKey == "" {
		translatorToEng := googletrans.NewGoogleTranslator("auto", "en")
		inputText, err = translatorToEng.Translate(inputText)
		if err != nil {
			return nil,errors.New(fmt.Sprintf("Failed to translate input text to English: %v", err))
		}
	} else if b.Language != "" && !contains(ALLOWED_LANGUAGES, b.Language) && b.GoogleTranslatorAPIKey != "" {
		inputText, err = googleOfficialTranslator.Translate(context.Background(), inputText, language.English, nil)
		if err != nil {
			return nil,errors.New(fmt.Sprintf("Failed to translate input text to English: %v", err))
		}
	}

	// Make post data structure and insert prompt
	inputTextStruct := [][]string{{inputText}, nil, {b.ConversationId, b.ResponseId, b.ChoiceId}}
	data := map[string]interface{}{
		"f.req": json.Marshal([]interface{}{nil, json.Marshal(inputTextStruct)}),
		"at":    b.SNlM0e,
	}

	// Get response
	resp, err := b.Session.Post("https://bard.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate", 
	params, data, b.Timeout, b.Proxies)
	if err != nil {
		return nil,errors.New(fmt.Sprintf("Failed to make POST request: %v", err))
	}

	// Post-processing of response
	respDict := make(map[string]interface{})
	err = json.Unmarshal([]byte(strings.Split(string(resp.Content), "\n")[3]), &respDict)
	if err != nil {
		return nil,errors.New(fmt.Sprintf("Failed to parse response: %v", err))
	}

	if len(respDict) == 0 {
		return map[string]interface{}{
			"content": fmt.Sprintf("Response Error: %s. \nTemporarily unavailable due to traffic or an error in cookie values. Please double-check the cookie values and verify your network environment.", resp.Content),
		}, nil
	}

	respJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(respDict["2"].(string)), &respJSON)
	if err != nil {
		return nil,errors.New(fmt.Sprintf("Failed to parse response JSON: %v", err))
	}

	// Gather image links (optional)
	images := make(map[string]interface{})
	if len(respJSON) >= 3 {
		nestedList := respJSON["4"].([]interface{})[0].([]interface{})[4].([]interface{})
		for _, img := range nestedList {
			images[img.([]interface{})[0].([]interface{})[0].(string)] = nil
		}
	}

	// Parsed Answer Object
	parsedAnswer := make(map[string]interface{})
	err = json.Unmarshal([]byte(respDict["2"].(string)), &parsedAnswer)
	if err != nil {
		return nil,errors.New(fmt.Sprintf("Failed to parse parsed answer: %v", err))
	}

	// Translated by Google Translator (optional)
	// Unofficial for testing
	if b.Language != "" && !contains(ALLOWED_LANGUAGES, b.Language) && b.GoogleTranslatorAPIKey == "" {
		translatorToLang := googletrans.NewGoogleTranslator("auto", b.Language)
		for i, x := range parsedAnswer["4"].([]interface{}) {
			parsedAnswer["4"].([]interface{})[i] = []interface{}{
				x.([]interface{})[0],
				append([]interface{}{translatorToLang.Translate(x.([]interface{})[1].(string))}, x.([]interface{})[1:]...),
				x.([]interface{})[2],
			}
		}
	} else if b.Language != "" && !contains(ALLOWED_LANGUAGES, b.Language) && b.GoogleTranslatorAPIKey != "" {
		for i, x := range parsedAnswer["4"].([]interface{}) {
			parsedAnswer["4"].([]interface{})[i] = []interface{}{
				x.([]interface{})[0],
				append([]interface{}{googleOfficialTranslator.Translate(context.Background(), x.([]interface{})[1].(string), language.English, nil)}, x.([]interface{})[1:]...),
				x.([]interface{})[2],
			}
		}
	}

	// Get code (optional)
	var code string
	if len(parsedAnswer["4"].([]interface{})) > 0 {
		code = strings.Split(parsedAnswer["4"].([]interface{})[0].([]interface{})[1].(string), "```")[1][6:]
	}

	// Returned dictionary object
	bardAnswer := map[string]interface{}{
		"content":           parsedAnswer["4"].([]interface{})[0].([]interface{})[1].(string),
		"conversation_id":   parsedAnswer["1"].([]interface{})[0],
		"response_id":       parsedAnswer["1"].([]interface{})[1],
		"factualityQueries": parsedAnswer["3"],
		"textQuery":         parsedAnswer["2"].([]interface{})[0],
		"choices": func() []map[string]interface{} {
			choices := make([]map[string]interface{}, len(parsedAnswer["4"].([]interface{})))
			for i, x := range parsedAnswer["4"].([]interface{}) {
				choices[i] = map[string]interface{}{
					"id":      x.([]interface{})[0],
					"content": x.([]interface{})[1],
				}
			}
			return choices
		}(),
		"links":  extractLinks(parsedAnswer["4"]),
		"images": images,
		"code":   code,
	}
	b.ConversationId, b.ResponseId, b.ChoiceId = bardAnswer["conversation_id"].(string), bardAnswer["response_id"].(string), bardAnswer["choices"].([]map[string]interface{})[0]["id"].(string)
	b.ReqId += 100000

	// Execute Code
	if b.RunCode && bardAnswer["code"] != nil {
		fmt.Println(bardAnswer["code"])
		_, err = govaluate.NewEvaluableExpression(bardAnswer["code"].(string)).Evaluate(nil)
		if err != nil {
			return nil,errors.New(fmt.Sprintf("Failed to execute code: %v", err))
		}
	}

	return bardAnswer,nil
}

func (b *Bard) Speech(inputText string, lang string) []byte {
	// Make POST request and return audio bytes
	// ...
	params := map[string]string{
		"bl":     "boq_assistant-bard-web-server_20230419.00_p1",
		"_reqid": strconv.Itoa(b.ReqId),
		"rt":     "c",
	}

	inputTextStruct := [][][]interface{}{
		{{"XqA3Ic", json.Marshal([]interface{}{nil, inputText, lang, nil, 2})}},
	}

	data := map[string]interface{}{
		"f.req": json.Marshal(inputTextStruct),
		"at":    b.SNlM0e,
	}

	// Get response
	resp, err := b.Session.Post(
		"https://bard.google.com/_/BardChatUi/data/batchexecute",
		params,
		data,
		b.Timeout,
		b.Proxies,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to make POST request: %v", err))
	}

	// Post-processing of response
	respDict := make([]interface{}, 0)
	err = json.Unmarshal([]byte(strings.Split(string(resp.Content), "\n")[3]), &respDict)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse response: %v", err))
	}

	if len(respDict) == 0 {
		return map[string]interface{}{
			"content": fmt.Sprintf("Response Error: %s. \nTemporarily unavailable due to traffic or an error in cookie values. Please double-check the cookie values and verify your network environment.", resp.Content),
		}
	}

	respJSON := make([]interface{}, 0)
	err = json.Unmarshal([]byte(respDict[0].([]interface{})[2].(string)), &respJSON)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse response JSON: %v", err))
	}

	audioB64 := respJSON[0].(string)
	audioBytes, err := base64.StdEncoding.DecodeString(audioB64)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode audio bytes: %v", err))
	}

	return audioBytes
}

func (b *Bard) uploadImage(image []byte, filename string) string {
	// Upload image into bard bucket on Google API
	// ...
	resp, err := http.Options("https://content-push.googleapis.com/upload/", nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to make OPTIONS request: %v", err))
	}
	resp.Body.Close()

	size := len(image)

	headers := IMG_UPLOAD_HEADERS
	headers.Set("size", strconv.Itoa(size))

	data := "File name: " + filename
	resp, err = http.Post("https://content-push.googleapis.com/upload/", "text/plain", strings.NewReader(data))
	if err != nil {
		panic(fmt.Sprintf("Failed to make POST request: %v", err))
	}
	resp.Body.Close()
	uploadURL := resp.Header.Get("X-Goog-Upload-Url")

	resp, err = http.Options(uploadURL, headers)
	if err != nil {
		panic(fmt.Sprintf("Failed to make OPTIONS request: %v", err))
	}
	resp.Body.Close()
	headers.Set("x-goog-upload-command", "upload, finalize")
	headers.Set("X-Goog-Upload-Offset", "0")

	resp, err = http.Post(uploadURL, "application/octet-stream", bytes.NewReader(image))
	if err != nil {
		panic(fmt.Sprintf("Failed to make POST request: %v", err))
	}
	resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("Failed to read response body")
	}

	return string(body)
}


func (b *Bard) askAboutImage(inputText string, image []byte, lang string, filename string) map[string]interface{} {
	// Send Bard image along with question and get answer
	// ...

	// Upload image
	imageURL := b.uploadImage(image, filename)

	inputDataStruct := []interface{}{
		nil,
		[]interface{}{
			[]interface{}{
				inputText,
				0,
				nil,
				[]interface{}{
					[]interface{}{
						imageURL,
						1,
					},
					filename,
				},
			},
			[]interface{}{lang},
			[]interface{}{"", "", ""},
			"", // Unknown random string value (1000 characters +)
			uuid.New().String(), // should be random uuidv4 (32 characters)
			nil,
			[]interface{}{1},
			0,
			[]interface{}{},
			[]interface{}{},
		},
	}

	params := map[string]string{
		"bl":     "boq_assistant-bard-web-server_20230419.00_p1",
		"_reqid": strconv.Itoa(b.ReqId),
		"rt":     "c",
	}

	inputDataStruct[1] = json.Marshal(inputDataStruct[1])
	data := map[string]interface{}{
		"f.req": json.Marshal(inputDataStruct),
		"at":    b.SNlM0e,
	}

	resp, err := b.Session.Post(
		"https://bard.google.com/u/1/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate",
		params,
		data,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to make POST request: %v", err))
	}

	// Post-processing of response
	respDict := make([]interface{}, 0)
	err = json.Unmarshal([]byte(strings.Split(string(resp.Content), "\n")[3]), &respDict)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse response: %v", err))
	}

	if len(respDict) == 0 {
		return map[string]interface{}{
			"content": fmt.Sprintf("Response Error: %s. \nTemporarily unavailable due to traffic or an error in cookie values. Please double-check the cookie values and verify your network environment.", resp.Content),
		}
	}

	parsedAnswer := make([]interface{}, 0)
	err = json.Unmarshal([]byte(respDict[0].([]interface{})[2].(string)), &parsedAnswer)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse response JSON: %v", err))
	}

	// Returned dictionary object
	bardAnswer := map[string]interface{}{
		"content":           parsedAnswer[4].([]interface{})[0].([]interface{})[1].(string),
		"conversation_id":   parsedAnswer[1].([]interface{})[0],
		"response_id":       parsedAnswer[1].([]interface{})[1],
		"factualityQueries": parsedAnswer[3],
		"textQuery":         parsedAnswer[2].([]interface{})[0],
		"choices": func() []map[string]interface{} {
			choices := make([]map[string]interface{}, len(parsedAnswer[4].([]interface{})))
			for i, x := range parsedAnswer[4].([]interface{}) {
				choices[i] = map[string]interface{}{
					"id":      x.([]interface{})[0],
					"content": x.([]interface{})[1],
				}
			}
			return choices
		}(),
		"links":  b.extractLinks(parsedAnswer[4]),
		"images": []string{""},
		"code":   "",
	}
	b.ConversationId, b.ResponseId, b.ChoiceId = bardAnswer["conversation_id"].(string), bardAnswer["response_id"].(string), bardAnswer["choices"].([]map[string]interface{})[0]["id"].(string)
	b.ReqId += 100000
	return bardAnswer
}


func (b *Bard) exportConversation(bardAnswer map[string]interface{}, title string) string {
	// Get Share URL for specific answer from bard
	// ...

	convID := bardAnswer["conversation_id"].(string)
	respID := bardAnswer["response_id"].(string)
	choiceID := bardAnswer["choices"].([]map[string]interface{})[0]["id"].(string)

	params := map[string]string{
		"rpcids":       "fuVx7",
		"source-path":  "/",
		"bl":           "boq_assistant-bard-web-server_20230713.13_p0",
		"rt":           "c",
	}

	inputDataStruct := [][][]interface{}{
		{{"fuVx7", json.Marshal([]interface{}{nil, []interface{}{[]interface{}{[]interface{}{convID, respID}, nil, nil, []interface{}{[]interface{}{}, []interface{}{}, []interface{}{}, choiceID, []interface{}{}}}, []interface{}{0, title}}), nil, "generic"})}},
	}

	data := map[string]interface{}{
		"f.req": json.Marshal(inputDataStruct),
		"at":    b.SNlM0e,
	}

	resp, err := b.Session.Post(
		"https://bard.google.com/_/BardChatUi/data/batchexecute",
		params,
		data,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to make POST request: %v", err))
	}

	// Post-processing of response
	respDict := make([]interface{}, 0)
	err = json.Unmarshal([]byte(strings.Split(string(resp.Content), "\n")[3]), &respDict)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse response: %v", err))
	}

	urlID := json.Unmarshal([]byte(respDict[0].([]interface{})[2].(string)))[2]
	url := fmt.Sprintf("https://g.co/bard/share/%s", urlID)

	// increment request ID
	b.ReqId += 100000

	return url
}
		
func (b *Bard) extractLinks(data interface{}) []string {
	links := []string{}
	if dataList, ok := data.([]interface{}); ok {
		for _, item := range dataList {
			if itemList, ok := item.([]interface{}); ok {
				links = append(links, b.extractLinks(itemList)...)
			} else if str, ok := item.(string); ok && strings.HasPrefix(str, "http") && !strings.Contains(str, "favicon") {
				links = append(links, str)
			}
		}
	}
	return links
}

func (b *Bard) ExtractCookie() string {
	// Extract __Secure-1PSID cookie from browsers
	// ...
	browsers := []func(string) ([]*http.Cookie, error){
		browsercookie.Chrome,
		browsercookie.Chromium,
		browsercookie.Opera,
		browsercookie.OperaGX,
		browsercookie.Brave,
		browsercookie.Edge,
		browsercookie.Vivaldi,
		browsercookie.Firefox,
		browsercookie.Librewolf,
		browsercookie.Safari,
	}
	for _, browserFn := range browsers {
		cookies, err := browserFn(".google.com")
		if err != nil {
			panic(fmt.Sprintf("Failed to get cookies from browser: %v", err))
		}
		for _, cookie := range cookies {
			if cookie.Name == "__Secure-1PSID" && strings.HasSuffix(cookie.Value, ".") {
				return cookie.Value
			}
		}
	}
	return ""
}

func (b *Bard) GetSNlM0e() string {
	// Extract __Secure-1PSID cookie from browsers
	// ...

	if b.Token == "" || b.Token[len(b.Token)-1] != '.' {
		panic("__Secure-1PSID value must end with a single dot. Enter correct __Secure-1PSID value.")
	}

	resp, err := b.Session.Get("https://bard.google.com/")
	if err != nil {
		panic(fmt.Sprintf("Response code not 200. Response Status is %d", resp.StatusCode))
	}

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("Response code not 200. Response Status is %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("Failed to read response body")
	}

	snlM0e := regexp.MustCompile(`SNlM0e":"(.*?)"`).FindStringSubmatch(string(body))
	if len(snlM0e) == 0 {
		panic("SNlM0e value not found. Double-check __Secure-1PSID value or pass it as token='xxxxx'.")
	}

	return snlM0e[1]

}

// func main() {
// 	bard := &Bard{
// 		Token: os.Getenv("_BARD_API_KEY"),
// 	}

// 	bardAnswer := bard.GetAnswer("Hello!")

// 	audio := bard.Speech("Tell me a joke.", "en-US")

// 	cookie := bard.ExtractCookie()
// }
