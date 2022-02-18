package tools

import (
	"bytes"
	"crypto/des"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/Zoxu0928/task-common/logger"
)

func MD5(str string) string {
	sha := md5.New()
	sha.Write([]byte(str))
	cipherStr := sha.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// Des加密
func DesEncrypt(key, text string) string {
	key = paddingKey(key, 8)
	bs, err := javaDesEncrypt([]byte(text), []byte(key))
	if err != nil {
		logger.Error("Encrypt des failed. Cause of %s.", err.Error())
		return ""
	}
	return hex.EncodeToString(bs)
}

// Des解密
func DesDecrypt(key, text string) string {
	key = paddingKey(key, 8)
	bytes, _ := hex.DecodeString(text)
	bs, err := javaDesDecrypt(bytes, []byte(key))
	if err != nil {
		logger.Error("Decrypt des failed. Cause of %s.", err.Error())
		return ""
	}
	return string(bs)
}

// java兼容的des加密
func javaDesEncrypt(origData, key []byte) ([]byte, error) {
	if len(key) > 8 {
		key = key[:8]
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	origData = pkcs5Padding(origData, bs)
	if len(origData)%bs != 0 {
		return nil, errors.New("Des encrypt failed. bytes length is wrong.")
	}

	out := make([]byte, len(origData))
	dst := out
	for len(origData) > 0 {
		block.Encrypt(dst, origData[:bs])
		origData = origData[bs:]
		dst = dst[bs:]
	}
	return out, nil

}

// java兼容的des解密
func javaDesDecrypt(crypted, key []byte) ([]byte, error) {
	if len(key) > 8 {
		key = key[:8]
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(crypted))
	dst := out
	bs := block.BlockSize()
	if len(crypted)%bs != 0 {
		return nil, errors.New("Des decrypt failed. bytes length is wrong.")
	}
	for len(crypted) > 0 {
		block.Decrypt(dst, crypted[:bs])
		crypted = crypted[bs:]
		dst = dst[bs:]
	}
	out = pkcs5UnPadding(out)
	return out, nil
}

// des处理
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// des处理
func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// 给字符串右补@号到一定长度
func paddingKey(key string, size int) string {
	for len(key) < size {
		key = key + "@"
	}
	return key
}
