package qiao

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// 私钥生成
// openssl genrsa -out rsa_private_key.pem 1024
var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDcGsUIIAINHfRTdMmgGwLrjzfMNSrtgIf4EGsNaYwmC1GjF/bM
h0Mcm10oLhNrKNYCTTQVGGIxuc5heKd1gOzb7bdTnCDPPZ7oV7p1B9Pud+6zPaco
qDz2M24vHFWYY2FbIIJh8fHhKcfXNXOLovdVBE7Zy682X1+R1lRK8D+vmQIDAQAB
AoGAeWAZvz1HZExca5k/hpbeqV+0+VtobMgwMs96+U53BpO/VRzl8Cu3CpNyb7HY
64L9YQ+J5QgpPhqkgIO0dMu/0RIXsmhvr2gcxmKObcqT3JQ6S4rjHTln49I2sYTz
7JEH4TcplKjSjHyq5MhHfA+CV2/AB2BO6G8limu7SheXuvECQQDwOpZrZDeTOOBk
z1vercawd+J9ll/FZYttnrWYTI1sSF1sNfZ7dUXPyYPQFZ0LQ1bhZGmWBZ6a6wd9
R+PKlmJvAkEA6o32c/WEXxW2zeh18sOO4wqUiBYq3L3hFObhcsUAY8jfykQefW8q
yPuuL02jLIajFWd0itjvIrzWnVmoUuXydwJAXGLrvllIVkIlah+lATprkypH3Gyc
YFnxCTNkOzIVoXMjGp6WMFylgIfLPZdSUiaPnxby1FNM7987fh7Lp/m12QJAK9iL
2JNtwkSR3p305oOuAz0oFORn8MnB+KFMRaMT9pNHWk0vke0lB1sc7ZTKyvkEJW0o
eQgic9DvIYzwDUcU8wJAIkKROzuzLi9AvLnLUrSdI6998lmeYO9x7pwZPukz3era
zncjRK3pbVkv0KrKfczuJiRlZ7dUzVO0b6QJr8TRAA==
-----END RSA PRIVATE KEY-----
`)

// 公钥: 根据私钥生成
// openssl rsa -in rsa_private_key.pem -pubout -out rsa_public_key.pem
var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDcGsUIIAINHfRTdMmgGwLrjzfM
NSrtgIf4EGsNaYwmC1GjF/bMh0Mcm10oLhNrKNYCTTQVGGIxuc5heKd1gOzb7bdT
nCDPPZ7oV7p1B9Pud+6zPacoqDz2M24vHFWYY2FbIIJh8fHhKcfXNXOLovdVBE7Z
y682X1+R1lRK8D+vmQIDAQAB
-----END PUBLIC KEY-----
`)

var block *pem.Block

// 加密
func RsaEncrypt(origData []byte) (Encrypt string, err error) {
	if block, _ = pem.Decode(publicKey); block == nil {
		return "", errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return
	}
	pub := pubInterface.(*rsa.PublicKey)
	EncryptByte, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	if err != nil {
		return
	}
	Encrypt = base64.StdEncoding.EncodeToString(EncryptByte)
	return
}

// 解密
func RsaDecrypt(ciphertext []byte) (Decrypt string, err error) {
	if block, _ = pem.Decode(privateKey); block == nil {
		return "", errors.New("private key error！")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return
	}
	DecryptByte, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	if err != nil {
		return
	}
	return string(DecryptByte), nil
}
