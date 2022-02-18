package tools

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"math/big"
	mathrandom "math/rand"
	"time"
)

func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// 获得UUID
func GetGuid() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

// 获得n以内的随机数
func GetRandomNum(n int) int {
	for {
		if random, err := rand.Int(rand.Reader, big.NewInt(int64(n))); err != nil {
			mathrandom.Seed(time.Now().UnixNano())
			continue
		} else {
			return int(random.Int64())
		}
	}
}
