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

var revokedTokens = make(map[string]time.Time)

var authList = make(map[string]*Auth)

func DefaultAuth(issuer string, aExp, rExp time.Duration, key string) *Auth {
	if authList[issuer] != nil {
		return authList[issuer]
	}
	authList[issuer] = &Auth{
		key:  []byte(key),
		aExp: aExp,
		rExp: rExp,
	}
	return authList[issuer]
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
func CreateClaims(issuer string, exp time.Duration) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    issuer,
		ExpiresAt: getJWTTime(exp),
	}
}

// NewDefaultToken /**
func (a *Auth) NewDefaultToken(data any) (aToken, rToken string, err error) {
	aToken, err = CreateToken(data, CreateClaims(a.issuer, a.aExp), a.key)
	if err != nil {
		return
	}
	rToken, err = CreateToken(nil, CreateClaims(a.issuer, a.rExp), a.key)
	if err != nil {
		return
	}
	return
}

// NewToken /**
func (a *Auth) NewToken(data any) (string, error) {
	return CreateToken(data, CreateClaims(a.issuer, a.aExp), a.key)
}

// CheckToken /**
func (a *Auth) CheckToken(token string) (any, error) {
	claims, err := VerifyToken(token, a.key)
	if err != nil {
		return nil, err
	}
	return claims.UserInfo, nil
}

// RefreshToken 刷新Token
func (a *Auth) RefreshToken(accessToken, refreshToken string) (string, string, error) {
	if _, err := VerifyToken(refreshToken, a.key); err != nil {
		return "", "", err
	}
	if claims, err := VerifyToken(accessToken, a.key); err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return a.NewDefaultToken(claims.UserInfo)
		}
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

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
