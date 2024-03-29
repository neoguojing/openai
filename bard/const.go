package bard

import "strings"

// import "github.com/fatih/color"

var ALLOWED_LANGUAGES = []string{"en", "ko", "ja", "english", "korean", "japanese", "chinese"}
var DEFAULT_LANGUAGE = "en"
var SEPARATOR_LINE = strings.Repeat("=", 36)
var USER_PROMPT = "You: "

var SESSION_HEADERS = map[string]string{
	"Host":          "bard.google.com",
	"X-Same-Domain": "1",
	"User-Agent":    "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
	"Content-Type":  "application/x-www-form-urlencoded;charset=UTF-8",
	"Origin":        "https://bard.google.com",
	"Referer":       "https://bard.google.com/",
}

var IMG_UPLOAD_HEADERS = map[string]string{
	"authority":                           "content-push.googleapis.com",
	"accept":                              "*/*",
	"accept-language":                     "en-US,en;q=0.7",
	"authorization":                       "Basic c2F2ZXM6cyNMdGhlNmxzd2F2b0RsN3J1d1U=", // constant authorization key
	"content-type":                        "application/x-www-form-urlencoded;charset=UTF-8",
	"origin":                              "https://bard.google.com",
	"push-id":                             "feeds/mcudyrk2a4khkz", // constant
	"referer":                             "https://bard.google.com/",
	"x-goog-upload-command":               "start",
	"x-goog-upload-header-content-length": "",
	"x-goog-upload-protocol":              "resumable",
	"x-tenant-id":                         "bard-storage",
}
