package openai

import (
	"encoding/json"

	"time"

	"github.com/go-resty/resty/v2"
	"github.com/neoguojing/log"
)

// BaiduResponse represents the response from ERNIE Bot
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

func (bc *BaiduClient) Complete(text string) (*ChatResponse, error) {
	resp, err := bc.client.R().
		SetQueryParams(map[string]string{
			"access_token": bc.token,
		}).
		SetBody(BaiduRequest{
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

	var result BaiduResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}
	return ConvertBaiduToOpenai(&result), nil
}

func (bc *BaiduClient) Close() {
	close(bc.quit)
}
