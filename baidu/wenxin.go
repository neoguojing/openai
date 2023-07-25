package baidu

import (
	"encoding/json"

	"time"

	"github.com/go-resty/resty/v2"
	"github.com/neoguojing/log"
)

// ErnieBotResponse represents the response from ERNIE Bot
type ErnieBotResponse struct {
	ID               string                 `json:"id"`
	Object           string                 `json:"object"`
	Created          int64                  `json:"created"`
	Result           string                 `json:"result"`
	IsTruncated      bool                   `json:"is_truncated"`
	NeedClearHistory bool                   `json:"need_clear_history"`
	Usage            map[string]interface{} `json:"usage"`
}

// UserMessage represents the user message structure for the Baidu API
type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Update Request structure to use UserMessage
type Request struct {
	Messages []UserMessage `json:"messages"`
}
type BaiduClient struct {
	client      *resty.Client
	credentials Credentials
	token       string
	quit        chan struct{}
}

type Credentials struct {
	key    string
	secret string
}

func NewBaiduClient(key, secret string) *BaiduClient {
	credentials := Credentials{
		key:    key,
		secret: secret,
	}
	client := resty.New()
	obj := &BaiduClient{
		client:      client,
		credentials: credentials,
		quit:        make(chan struct{}),
	}
	obj.RefreshToken()
	return obj
}

// RefreshToken periodically refreshes the access token
func (bc *BaiduClient) RefreshToken() {
	bc.token = bc.GetAccessToken()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				bc.token = bc.GetAccessToken()
			case <-bc.quit:
				return
			}
		}
	}()
}

func (bc *BaiduClient) GetAccessToken() string {
	resp, err := bc.client.R().
		SetQueryParams(map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     bc.credentials.key,
			"client_secret": bc.credentials.secret,
		}).
		SetHeader("Accept", "application/json").
		Post("https://aip.baidubce.com/oauth/2.0/token")

	if err != nil {
		log.Error(err.Error())
		return ""
	}

	var result map[string]interface{}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		log.Error(err.Error())
		return ""
	}

	return result["access_token"].(string)
}

func (bc *BaiduClient) Complete(text string) (*ErnieBotResponse, error) {
	resp, err := bc.client.R().
		SetQueryParams(map[string]string{
			"access_token": bc.token,
		}).
		SetBody(Request{
			Messages: []UserMessage{
				{
					Role:    "user",
					Content: text,
				},
			},
		}).
		SetHeader("Content-Type", "application/json").
		Post("https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant")

	if err != nil {
		return nil, err
	}

	var result ErnieBotResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (bc *BaiduClient) Close() {
	close(bc.quit)
}
