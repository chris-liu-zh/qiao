package jwt

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type NumericDate struct {
	time.Time
}

func NewNumericDate(t time.Time) *NumericDate {
	return &NumericDate{t.Truncate(time.Second)}
}

type Claims interface {
	GetExpirationTime() *NumericDate // 获取过期时间
	GetNotBefore() *NumericDate      // 获取生效时间
	GetIssuedAt() *NumericDate       // 获取签发时间
	GetIssuer() string               // 获取签发人
	GetSubject() string              // 获取主题
	GetAudience() []string           // 获取受众
	GetID() string                   // 获取ID
}

type RegisteredClaims struct {
	Issuer    string       `json:"iss,omitempty"` // 签发人
	Subject   string       `json:"sub,omitempty"` // 主题
	Audience  []string     `json:"aud,omitempty"` // 受众
	ExpiresAt *NumericDate `json:"exp,omitempty"` // 过期时间
	NotBefore *NumericDate `json:"nbf,omitempty"` // 生效时间
	IssuedAt  *NumericDate `json:"iat,omitempty"` // 签发时间
	ID        string       `json:"jti,omitempty"` // ID
}

func (c RegisteredClaims) GetExpirationTime() *NumericDate {
	return c.ExpiresAt
}

func (c RegisteredClaims) GetNotBefore() *NumericDate {
	return c.NotBefore
}

func (c RegisteredClaims) GetIssuedAt() *NumericDate {
	return c.IssuedAt
}

func (c RegisteredClaims) GetIssuer() string {
	return c.Issuer
}

func (c RegisteredClaims) GetSubject() string {
	return c.Subject
}

func (c RegisteredClaims) GetAudience() []string {
	return c.Audience
}
func (c RegisteredClaims) GetID() string {
	return c.ID
}

func Base64Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func Base64Decode(base64Str string, v any) error {
	data, err := base64.RawURLEncoding.DecodeString(base64Str)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}
