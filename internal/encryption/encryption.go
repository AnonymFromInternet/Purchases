package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type Encryption struct {
	Key []byte
}

func (encryption *Encryption) Encrypt(text string) (encryptedText string, err error) {
	textAsBytes := []byte(text)

	block, err := aes.NewCipher(encryption.Key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(textAsBytes))
	iv := cipherText[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], textAsBytes)

	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func (encryption *Encryption) Decrypt(encryptedText string) (decryptedText string, error error) {
	cipherText, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryption.Key)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("length of the cipherText is smaller than aes.BlockSize")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return fmt.Sprintf("%s", cipherText), nil

}
