package bard

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/tebeka/selenium"
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
	ResponseId             string
	ChoiceId               string
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

	if b.Token == "" && token == "" {
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
		b.Session.Jar.SetCookies(&url.URL{Host: "https://bard.google.com"}, []*http.Cookie{cookie})
	}

	b.SNlM0e = b.GetSNlM0e()

	return b
}

func (b *Bard) GetAnswer(inputText string) (map[string]interface{}, error) {
	// Make POST request and parse response
	// ...
	// if b.GoogleTranslatorAPIKey != "" {
	// 	googleOfficialTranslator, err := translate.NewClient(context.Background(), option.WithAPIKey(b.GoogleTranslatorAPIKey))
	// 	if err != nil {
	// 		return nil,errors.New(fmt.Sprintf("Failed to create Google Translator client: %v", err))
	// 	}
	// }

	// Set language (optional)
	// if b.Language != "" && !contains(ALLOWED_LANGUAGES, b.Language) && b.GoogleTranslatorAPIKey == "" {
	// 	translatorToEng := googletrans.NewGoogleTranslator("auto", "en")
	// 	inputText, err = translatorToEng.Translate(inputText)
	// 	if err != nil {
	// 		return nil,errors.New(fmt.Sprintf("Failed to translate input text to English: %v", err))
	// 	}
	// } else if b.Language != "" && !contains(ALLOWED_LANGUAGES, b.Language) && b.GoogleTranslatorAPIKey != "" {
	// 	inputText, err = googleOfficialTranslator.Translate(context.Background(), inputText, language.English, nil)
	// 	if err != nil {
	// 		return nil,errors.New(fmt.Sprintf("Failed to translate input text to English: %v", err))
	// 	}
	// }

	params := url.Values{}
	params.Set("bl", "boq_assistant-bard-web-server_20230419.00_p1")
	params.Set("_reqid", fmt.Sprint(b.ReqId))
	params.Set("rt", "c")
	reqURL := "https://bard.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate?" + params.Encode()
	// Make post data structure and insert prompt
	inputTextStruct := [][]string{{inputText}, nil, {b.ConversationId, b.ResponseId, b.ChoiceId}}
	elem, _ := json.Marshal(inputTextStruct)
	req, _ := json.Marshal([]interface{}{nil, elem})
	data := map[string]interface{}{
		"f.req": req,
		"at":    b.SNlM0e,
	}

	dataBytes, _ := json.Marshal(data)
	reqBody := bytes.NewBuffer(dataBytes)
	// Get response
	resp, err := b.Session.Post(reqURL, "", reqBody)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to make POST request: %v", err))
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to read response body: %v", err))
	}

	// Post-processing of response
	respDict := make(map[string]interface{})
	err = json.Unmarshal(respBody, &respDict)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to parse response: %v", err))
	}

	if len(respDict) == 0 {
		return map[string]interface{}{
			"content": fmt.Sprintf("Response Error: %s. \nTemporarily unavailable due to traffic or an error in cookie values. Please double-check the cookie values and verify your network environment."),
		}, nil
	}

	respJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(respDict["2"].(string)), &respJSON)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to parse response JSON: %v", err))
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
		return nil, errors.New(fmt.Sprintf("Failed to parse parsed answer: %v", err))
	}

	// Translated by Google Translator (optional)
	// Unofficial for testing
	// if b.Language != "" && !contains(ALLOWED_LANGUAGES, b.Language) && b.GoogleTranslatorAPIKey == "" {
	// 	translatorToLang := googletrans.NewGoogleTranslator("auto", b.Language)
	// 	for i, x := range parsedAnswer["4"].([]interface{}) {
	// 		parsedAnswer["4"].([]interface{})[i] = []interface{}{
	// 			x.([]interface{})[0],
	// 			append([]interface{}{translatorToLang.Translate(x.([]interface{})[1].(string))}, x.([]interface{})[1:]...),
	// 			x.([]interface{})[2],
	// 		}
	// 	}
	// } else if b.Language != "" && !contains(ALLOWED_LANGUAGES, b.Language) && b.GoogleTranslatorAPIKey != "" {
	// 	for i, x := range parsedAnswer["4"].([]interface{}) {
	// 		parsedAnswer["4"].([]interface{})[i] = []interface{}{
	// 			x.([]interface{})[0],
	// 			append([]interface{}{googleOfficialTranslator.Translate(context.Background(), x.([]interface{})[1].(string), language.English, nil)}, x.([]interface{})[1:]...),
	// 			x.([]interface{})[2],
	// 		}
	// 	}
	// }

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
		"links":  b.extractLinks(parsedAnswer["4"]),
		"images": images,
		"code":   code,
	}
	b.ConversationId, b.ResponseId, b.ChoiceId = bardAnswer["conversation_id"].(string), bardAnswer["response_id"].(string),
		bardAnswer["choices"].([]map[string]interface{})[0]["id"].(string)
	b.ReqId += 100000

	// Execute Code
	if b.RunCode && bardAnswer["code"] != nil {
		fmt.Println(bardAnswer["code"])
	}

	return bardAnswer, nil
}

func (b *Bard) Speech(inputText string, lang string) ([]byte, error) {
	// Make POST request and return audio bytes
	// ...

	inputElem, _ := json.Marshal([]interface{}{nil, inputText, lang, nil, 2})
	inputTextStruct := [][][]interface{}{
		{{"XqA3Ic", inputElem}},
	}

	req, _ := json.Marshal(inputTextStruct)

	data := map[string]interface{}{
		"f.req": req,
		"at":    b.SNlM0e,
	}

	params := url.Values{}
	params.Set("bl", "boq_assistant-bard-web-server_20230419.00_p1")
	params.Set("_reqid", strconv.Itoa(b.ReqId))
	params.Set("rt", "c")
	reqURL := "https://bard.google.com/_/BardChatUi/data/batchexecute?" + params.Encode()

	dataBytes, _ := json.Marshal(data)
	reqBody := bytes.NewBuffer(dataBytes)
	// Get response
	resp, err := b.Session.Post(reqURL, "", reqBody)
	if err != nil {
		panic(fmt.Sprintf("Failed to make POST request: %v", err))
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to read response body: %v", err))
	}

	// Post-processing of response
	respDict := make([]interface{}, 0)
	err = json.Unmarshal(respBody, &respDict)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse response: %v", err))
	}

	if len(respDict) == 0 {
		return nil, errors.New("respDict empty")
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

	return audioBytes, nil
}

func (b *Bard) uploadImage(image []byte, filename string) string {
	// Upload image into bard bucket on Google API
	// ...
	client := &http.Client{}
	req, err := http.NewRequest("OPTIONS", "https://content-push.googleapis.com/upload/", nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to make OPTIONS request: %v", err))
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	data := "File name: " + filename
	req, err = http.NewRequest("POST", "https://content-push.googleapis.com/upload/", strings.NewReader(data))
	size := len(image)

	req.Header.Set("size", strconv.Itoa(size))
	resp, err = client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Failed to make POST request: %v", err))
	}
	resp.Body.Close()
	uploadURL := resp.Header.Get("X-Goog-Upload-Url")

	req, err = http.NewRequest("OPTIONS", uploadURL, nil)
	resp, err = client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Failed to make OPTIONS request: %v", err))
	}
	resp.Body.Close()

	req, err = http.NewRequest("POST", uploadURL, bytes.NewReader(image))
	req.Header.Set("x-goog-upload-command", "upload, finalize")
	req.Header.Set("X-Goog-Upload-Offset", "0")
	resp, err = client.Do(req)
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
			"",                  // Unknown random string value (1000 characters +)
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
		"rpcids":      "fuVx7",
		"source-path": "/",
		"bl":          "boq_assistant-bard-web-server_20230713.13_p0",
		"rt":          "c",
	}

	elem, _ := json.Marshal([]interface{}{
		nil,
		[]interface{}{
			[]interface{}{
				[]interface{}{convID, respID},
				nil, nil,
				[]interface{}{[]interface{}{}, []interface{}{}, []interface{}{}, choiceID, []interface{}{}},
			},
			[]interface{}{0, title},
		},
	})
	inputDataStruct := [][][]interface{}{
		{
			{
				"fuVx7",
				elem,
				nil,
				"generic",
			},
		},
	}

	req, _ := json.Marshal(inputDataStruct)
	data := map[string]interface{}{
		"f.req": req,
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
		getCookieFromChrome,
		getCookieFromFirfox,
	}

	for _, browserFn := range browsers {
		cookies, err := browserFn("www.google.com")
		if err != nil {
			fmt.Printf("Failed to get cookies from browser: %v", err)
			continue
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

func getCookieFromChrome(url string) ([]*http.Cookie, error) {
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		return nil, err
	}
	defer wd.Quit()

	// Navigate to the page.
	if err := wd.Get(url); err != nil {
		return nil, err
	}

	// Get a cookie.
	cookie, err := wd.GetCookies()
	if err != nil {
		return nil, err
	}
	return cookie
}

func getCookieFromFirfox(url, string) ([]*http.Cookie, error) {
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		return nil, err
	}
	defer wd.Quit()

	// Navigate to the page.
	if err := wd.Get(url); err != nil {
		return nil, err
	}

	// Get a cookie.
	cookie, err := wd.GetCookies()
	if err != nil {
		return nil, err
	}
	return cookie
}

// func main() {
// 	bard := &Bard{
// 		Token: os.Getenv("_BARD_API_KEY"),
// 	}

// 	bardAnswer := bard.GetAnswer("Hello!")

// 	audio := bard.Speech("Tell me a joke.", "en-US")

// 	cookie := bard.ExtractCookie()
// }
