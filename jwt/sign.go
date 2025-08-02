package jwt

import (
	"encoding/json"
	"errors"
	"strings"
)

type Header struct {
	Typ string `json:"typ"` // 类型
	Alg string `json:"alg"` // 签名算法
}

type Token struct {
	method    *SigningMethod
	header    *Header
	Claims    Claims
	signature []byte
}

func NewToken(method *SigningMethod, claims Claims) *Token {
	return &Token{
		header: &Header{
			Typ: "JWT",           // 默认使用JWT类型
			Alg: method.GetAlg(), // 默认使用HS256算法
		},
		Claims: claims,
		method: method,
	}
}

func (t *Token) signString() (string, error) {
	if t.Claims == nil {
		return "", errors.New("header or payload is nil")
	}
	h, err := json.Marshal(t.header)
	if err != nil {
		return "", err
	}
	p, err := json.Marshal(t.Claims)
	if err != nil {
		return "", err
	}
	return Base64Encode(h) + "." + Base64Encode(p), nil
}

func (t *Token) Sign(key []byte) (string, error) {
	signingString, err := t.signString()
	if err != nil {
		return "", err
	}
	t.signature = t.method.Sign(signingString, key)
	return signingString + "." + Base64Encode(t.signature), nil
}

func ParseWithClaims(token string, claims Claims, key []byte) (t *Token, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}
	headerBase64 := parts[0]
	payloadBase64 := parts[1]
	signatureBase64 := parts[2]
	t = &Token{}
	h := &Header{}
	if err := Base64Decode(headerBase64, h); err != nil {
		return nil, err
	}
	if err := Base64Decode(payloadBase64, &claims); err != nil {
		return nil, err
	}
	t.header = h
	t.Claims = claims
	t.method = GetSigningMethod(t.header.Alg)
	if t.method == nil {
		return nil, errors.New("invalid algorithm")
	}
	if signatureBase64 != Base64Encode(t.method.Sign(headerBase64+"."+payloadBase64, key)) {
		return nil, errors.New("invalid signature")
	}
	return t, nil
}
