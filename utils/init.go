package utils

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	slog "github.com/vearne/simplelog"
	"io"
	"os"
)

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func Sha256N(plaintext string, n int) string {
	h := sha256.New()
	buff := []byte(plaintext)
	for i := 0; i < n; i++ {
		_, err := h.Write(buff)
		if err != nil {
			slog.Error("Sha256N:%v error", err)
		}
		buff = h.Sum(nil)
		h.Reset()
	}
	return fmt.Sprintf("%x", buff)
}

func GenRandIV() []byte {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		slog.Error("GenRandIV:%v error", err)
	}
	return iv
}

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func FindInSlice(keyword string, slice []string) bool {
	for _, item := range slice {
		if keyword == item {
			return true
		}
	}
	return false
}

func CalcTotalPage(total, pageSize int) int {
	x := total / pageSize
	if total%pageSize != 0 {
		x++
	}
	return x
}

func IsSecurePassword(s string) bool {
	if len(s) < 8 {
		return false
	}
	var hasLowerCaseChar bool
	var hasUpperCaseChar bool
	var hasNumberChar bool
	var hasSpecialChar bool
	for _, char := range s {
		if char >= 'a' && char <= 'z' {
			hasLowerCaseChar = true
		} else if char >= 'A' && char <= 'Z' {
			hasUpperCaseChar = true
		} else if char >= '0' && char <= '9' {
			hasNumberChar = true
		} else {
			hasSpecialChar = true
		}
	}
	return hasLowerCaseChar && hasUpperCaseChar && hasNumberChar && hasSpecialChar
}
