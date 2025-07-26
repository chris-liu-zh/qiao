package Http

/*
 * @Author: Chris
 * @Date: 2023-06-13 14:17:57
 * @LastEditors: Strong
 * @LastEditTime: 2025-03-22 16:17:39
 * @Description: 请填写简介
 */

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chris-liu-zh/qiao"

	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	issuer string
	key    []byte
	aExp   time.Duration
	rExp   time.Duration
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

var revokedTokens = make(map[string]time.Time)

var authList = make(map[string]*Auth)

func DefaultAuth(issuer string, aExp, rExp time.Duration, key string) (*Auth, error) {
	if authList[issuer] != nil {
		return nil, fmt.Errorf("auth already exists for issuer: %s", issuer)
	}
	authList[issuer] = &Auth{
		key:  []byte(key),
		aExp: aExp,
		rExp: rExp,
	}
	return authList[issuer], nil
}

func getAuth(issuer string) (*Auth, error) {
	auth := authList[issuer]
	if auth == nil {
		return nil, fmt.Errorf("auth not found for issuer: %s", issuer)
	}
	return auth, nil
}

// DefaultSign /**
func DefaultSign(sign, appKey, secret string, ts time.Time, timeDiff time.Duration) error {
	now := time.Now()
	// 检查时间戳是否在有效时间范围内
	if ts.Before(now.Add(-timeDiff)) || ts.After(now.Add(timeDiff)) {
		return errors.New("timestamp 超出有效时间范围，请检查系统时间")
	}
	s := fmt.Sprintf("%s%s%d", appKey, secret, ts.Unix())
	localSign := strings.ToUpper(qiao.MD5(s))

	if localSign != sign {
		return errors.New("sign error")
	}
	return nil
}

// CreateClaims 创建 JWT 注册声明
func NewClaims(issuer string) (*Claims, error) {
	a, err := getAuth(issuer)
	if err != nil {
		return nil, err
	}
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: a.issuer,
		},
	}, nil
}
func (c *Claims) SetExpiresAt(t time.Duration) *Claims {
	c.ExpiresAt = getJWTTime(t)
	return c
}

func (c *Claims) SetSubject(subject string) *Claims {
	c.Subject = subject
	return c
}

// SetIssuedAt 设置签发时间
func (c *Claims) SetIssuedAt() *Claims {
	c.IssuedAt = jwt.NewNumericDate(time.Now())
	return c
}

// SetNotBefore 设置不早于时间
func (c *Claims) SetNotBefore() *Claims {
	c.NotBefore = jwt.NewNumericDate(time.Now())
	return c
}

// SetAudience 设置受众
func (c *Claims) SetAudience(audience string) *Claims {
	if c.Audience == nil {
		c.Audience = make([]string, 0)
	}
	c.Audience = append(c.Audience, audience)
	return c
}

// 新建默认的双Token
func (c *Claims) NewDefaultToken() (t Token, err error) {
	a, err := getAuth(c.Issuer)
	if err != nil {
		return
	}
	if t.AccessToken, err = c.SetExpiresAt(a.aExp).CreateToken(a.key); err != nil {
		return
	}
	if t.RefreshToken, err = c.SetExpiresAt(a.rExp).CreateToken(a.key); err != nil {
		return
	}
	return
}

// NewToken 创建新的 Token
func (c *Claims) NewToken() (t Token, err error) {
	a, err := getAuth(c.Issuer)
	if err != nil {
		return
	}
	t.AccessToken, err = c.SetExpiresAt(a.aExp).CreateToken(a.key)
	if err != nil {
		return
	}
	return
}

// CheckToken 验证Token
func CheckToken(issuer string, token string) (subject string, err error) {
	a, err := getAuth(issuer)
	if err != nil {
		return
	}
	claims, err := VerifyToken(token, issuer, a.key)
	if err != nil {
		return
	}
	return claims.Subject, nil
}

// RefreshToken 刷新Token
func RefreshToken(accessToken, refreshToken, issuer string) (t Token, err error) {
	a, err := getAuth(issuer)
	if err != nil {
		return
	}
	refreshClaims, err := VerifyToken(refreshToken, issuer, a.key)
	if err != nil {
		return
	}
	claims, err := VerifyToken(accessToken, issuer, a.key)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return
	}
	if refreshClaims.Subject != claims.Subject {
		return t, errors.New("subject mismatch")
	}
	return refreshClaims.NewDefaultToken()
}

// RevokeToken 撤销过期无效的Token
func init() {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			now := time.Now()
			for token, expiresIn := range revokedTokens {
				if now.After(expiresIn) {
					delete(revokedTokens, token)
				}
			}
		}
	}()
}

// SetInvalidToken 将Token设置为无效
func SetInvalidToken(token string) error {
	claims := &Claims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return err
	}
	if claims.ExpiresAt != nil {
		revokedTokens[token] = claims.ExpiresAt.Time
	}
	return nil
}
