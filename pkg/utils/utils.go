package utils

import (
	"fmt"
	"math/rand"
	"net/url"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CreateMessageURL(account, messageSubject, code string) string {
	return fmt.Sprintf("https://www.reddit.com/message/compose/?to=%s&subject=%s&message=%s",
		url.QueryEscape(account),
		url.QueryEscape(messageSubject),
		url.QueryEscape(code),
	)
}
