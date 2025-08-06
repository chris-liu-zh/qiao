package jwt

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const tokenDelimiter = "."

type Header struct {
	Typ string `json:"typ"` // 类型
	Alg string `json:"alg"` // 签名算法
}

type Token struct {
	method    *SigningMethod
	header    Header
	Claims    Claims
	signature []byte
}

func NewToken(method *SigningMethod, claims Claims) *Token {
	return &Token{
		header: Header{
			Typ: "JWT",           // 默认使用JWT类型
			Alg: method.GetAlg(), // 默认使用HS256算法
		},
		Claims: claims,
		method: method,
	}
}

func (t *Token) signString() (string, error) {
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

func VerifyToken(tokenStr string, claims Claims, key []byte) error {
	t, err := stringToToken(tokenStr, claims, key)
	if err != nil {
		return err
	}
	timeNow := time.Now()
	if t.Claims.GetExpirationTime() != nil && t.Claims.GetExpirationTime().Time.Before(timeNow) {
		return ErrTokenExpired
	}
	if t.Claims.GetNotBefore() != nil && t.Claims.GetNotBefore().Time.After(timeNow) {
		return ErrTokenNotValidYet
	}
	return nil
}

func splitToken(token string) ([]string, bool) {
	parts := make([]string, 3)
	header, remain, ok := strings.Cut(token, tokenDelimiter)
	if !ok {
		return nil, false
	}
	parts[0] = header
	claims, remain, ok := strings.Cut(remain, tokenDelimiter)
	if !ok {
		return nil, false
	}
	parts[1] = claims
	signature, _, unexpected := strings.Cut(remain, tokenDelimiter)
	if unexpected {
		return nil, false
	}
	parts[2] = signature

	return parts, true
}

func stringToToken(token string, claims Claims, key []byte) (t *Token, err error) {
	parts, ok := splitToken(token)
	if !ok {
		return nil, ErrTokenMalformed
	}
	signingString := strings.Join(parts[0:2], ".")
	t = &Token{}
	if err := Base64Decode(parts[0], &t.header); err != nil {
		return nil, ErrInvalidHeader
	}
	if err := Base64Decode(parts[1], claims); err != nil {
		return nil, ErrTokenInvalidClaims
	}
	t.Claims = claims
	t.method = GetSigningMethod(t.header.Alg)
	if t.method == nil {
		return nil, ErrHashUnavailable
	}
	t.signature, err = base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrTokenSignatureInvalid
	}
	signature := t.method.Sign(signingString, key)
	if subtle.ConstantTimeCompare(t.signature, signature) != 1 {
		return nil, ErrTokenUnverifiable
	}
	return t, nil
}
