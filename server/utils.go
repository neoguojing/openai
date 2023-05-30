package main

import (
	"crypto/md5"
	"encoding/hex"
)

func GenerateUserIdentifier(userAgent, acceptLanguage, forwardedFor string) string {
	// 将User-Agent和Accept-Language标头中的值连接起来
	uaAndLang := userAgent + acceptLanguage

	// 如果有X-Forwarded-For标头，则将其添加到uaAndLang中
	if forwardedFor != "" {
		uaAndLang += forwardedFor
	}

	// 使用MD5哈希算法生成唯一标识符
	hash := md5.Sum([]byte(uaAndLang))
	return hex.EncodeToString(hash[:])
}
