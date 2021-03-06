package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	slog "github.com/vearne/simplelog"
)

var Space = "                "

func paddingSpace(plaintext []byte) []byte {
	mod := len(plaintext) % aes.BlockSize
	buff := bytes.NewBuffer(plaintext)
	if mod != 0 {
		buff.Write([]byte(Space)[0 : aes.BlockSize-mod])
	}
	return buff.Bytes()
}

func GenHMacKey(data []byte, salt []byte) []byte {
	h := hmac.New(sha256.New, salt)
	// Write Data to it
	_, err := h.Write(data)
	if err != nil {
		slog.Error("GenHMacKey, %v", err)
	}
	// Get result and encode as hexadecimal string
	return h.Sum(nil)
}

func EncryptAesInCFB(plaintext []byte, key []byte, iv []byte) []byte {
	plaintext = paddingSpace(plaintext)
	block, err := aes.NewCipher(key)
	if err != nil {
		return make([]byte, 0)
	}

	dst := make([]byte, len(plaintext))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(dst, plaintext)
	return dst
}

func DecryptAesInCFB(ciphertext []byte, key []byte, iv []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		return make([]byte, 0)
	}

	stream := cipher.NewCFBDecrypter(block, iv)

	dst := make([]byte, len(ciphertext))
	stream.XORKeyStream(dst, ciphertext)

	return bytes.TrimSpace(dst)
}
