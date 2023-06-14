package urlsigner

import (
	"fmt"
	goalone "github.com/bwmarrin/go-alone"
	"strings"
	"time"
)

type Signer struct {
	Secret []byte
}

func (signer *Signer) GenerateTokenFromString(s string) string {
	var urlToSign string

	crypt := goalone.New(signer.Secret, goalone.Timestamp)

	if strings.Contains(s, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", s)
	} else {
		urlToSign = fmt.Sprintf("%s?hash=", s)
	}

	tokenBytes := crypt.Sign([]byte(urlToSign))

	return string(tokenBytes)
}

func (signer *Signer) IsTokenValid(token string) bool {
	crypt := goalone.New(signer.Secret, goalone.Timestamp)
	_, err := crypt.Unsign([]byte(token))
	if err != nil {
		fmt.Println("cannot unsign the token :", err)
		return false
	}

	return true
}

func (signer *Signer) IsTokenActual(token string, minutesUntilExpire int) bool {
	crypt := goalone.New(signer.Secret, goalone.Timestamp)
	ts := crypt.Parse([]byte(token))

	return time.Since(ts.Timestamp) > time.Duration(minutesUntilExpire)*time.Minute
}
