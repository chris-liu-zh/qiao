package qiao

import (
	"crypto/aes"
	"encoding/base64"
	"errors"
)

/**
 * @description:加密
 * @param {*} encryptString
 * @param {string} aesKey
 * @return {*}
 */
func AESEncrypt(encryptString, aesKey string) (encrypted string, err error) {
	key := []byte(aesKey)
	src := []byte(encryptString)
	cipher, err := aes.NewCipher(generateKey(key))
	if err != nil {
		return
	}
	length := (len(src) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, src)
	pad := byte(len(plain) - len(src))
	for i := len(src); i < len(plain); i++ {
		plain[i] = pad
	}
	encryptedByte := make([]byte, len(plain))
	for bs, be := 0, cipher.BlockSize(); bs <= len(src); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encryptedByte[bs:be], plain[bs:be])
	}
	encrypted = base64.StdEncoding.EncodeToString(encryptedByte)
	return
}

// 解密
func AESDecrypt(decodeString, aesKey string) (decrypted string, err error) {
	key := []byte(aesKey)
	decryptCode, err := base64.StdEncoding.DecodeString(decodeString)
	if err != nil {
		return
	}
	cipher, err := aes.NewCipher(generateKey(key))
	if err != nil {
		return
	}
	decryptedByte := make([]byte, len(decryptCode))
	for bs, be := 0, cipher.BlockSize(); bs < len(decryptCode); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decryptedByte[bs:be], decryptCode[bs:be])
	}
	trim := 0
	if len(decryptedByte) > 0 {
		trim = len(decryptedByte) - int(decryptedByte[len(decryptedByte)-1])
	}
	if trim < 0 {
		return "", errors.New("密钥或加密数据不正确！")
	}
	return string(decryptedByte), nil
}

func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}
