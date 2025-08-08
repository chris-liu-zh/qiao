package jwt

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
)

type Auth struct {
	issuer      string
	key         []byte
	accessExp   time.Duration
	refreshrExp time.Duration
}

type DefaultToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

var (
	authList          = make(map[string]*Auth) // 认证列表
	ErrIssuerNotExist = errors.New("issuer not exist")
)

type ClaimsOption func(*DefaultClaims)

type DefaultClaims struct {
	RegisteredClaims
	UserInfo any    `json:"user_info,omitempty"`
	Version  string `json:"ver,omitempty"`
}

func WithSubject(sub string) ClaimsOption {
	return func(c *DefaultClaims) {
		c.Subject = sub
	}
}

func WithVersion(ver string) ClaimsOption {
	return func(c *DefaultClaims) {
		c.Version = ver
	}
}

func WithUserInfo(userInfo any) ClaimsOption {
	return func(c *DefaultClaims) {
		c.UserInfo = userInfo
	}
}

func getJwtDate(exp time.Duration) *NumericDate {
	if exp > 0 {
		return NewNumericDate(time.Now().Add(exp))
	}
	return NewNumericDate(time.Now())
}

func SetAuth(issuer string, accessExp, refreshrExp time.Duration, key string) error {
	if _, ok := authList[issuer]; ok {
		return errors.New("issuer already exists")
	}
	authList[issuer] = &Auth{
		issuer:      issuer,
		key:         []byte(key),
		accessExp:   accessExp,
		refreshrExp: refreshrExp,
	}
	return nil
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

// CreateToken 创建新的 DefaultToken
func CreateToken(issuer string, claimsOption ...ClaimsOption) (t DefaultToken, err error) {
	if auth, ok := authList[issuer]; ok {
		ac := DefaultClaims{
			RegisteredClaims: RegisteredClaims{
				ExpiresAt: getJwtDate(auth.accessExp),
				Issuer:    auth.issuer,
			},
		}
		for _, option := range claimsOption {
			option(&ac)
		}
		fmt.Println(ac)
		t.AccessToken, err = NewToken(SignMethodHS256, ac).Sign(auth.key)
		if err != nil {
			return
		}
		if auth.refreshrExp > 0 {
			ac.ExpiresAt = getJwtDate(auth.refreshrExp)
			t.RefreshToken, err = NewToken(SignMethodHS256, ac).Sign(auth.key)
			if err != nil {
				return
			}
		}
		return t, nil
	}
	return t, ErrIssuerNotExist
}

// CheckToken 验证Token
func CheckToken(issuer, token string, userinfo any) error {
	if auth, ok := authList[issuer]; ok {
		c := &DefaultClaims{
			UserInfo: userinfo,
		}
		if err := VerifyToken(token, c, auth.key); err != nil {
			return err
		}
		if c.Issuer != auth.issuer {
			return errors.New("token issuer error")
		}
		return nil
	}
	return ErrIssuerNotExist
}

func GetClaims(issuer, token string) (*DefaultClaims, error) {
	if auth, ok := authList[issuer]; ok {
		claims := &DefaultClaims{}
		if err := VerifyToken(token, claims, auth.key); err != nil {
			if errors.Is(err, ErrTokenExpired) {
				return claims, err
			}
			return nil, err
		}
		return claims, nil
	}
	return nil, ErrIssuerNotExist
}

// RefreshToken 刷新Token
func RefreshToken(issuer, accessToken, refreshrToken string) (t DefaultToken, err error) {
	accClaims, err := GetClaims(issuer, accessToken)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			refClaims, err := GetClaims(issuer, refreshrToken)
			if err != nil {
				return t, err
			}
			if refClaims.Subject != accClaims.Subject && refClaims.Issuer != accClaims.Issuer {
				return t, errors.New("refresh token error")
			}
			auth := authList[issuer]
			accClaims.ExpiresAt = getJwtDate(auth.accessExp)
			refClaims.ExpiresAt = getJwtDate(auth.refreshrExp)
			if t.AccessToken, err = NewToken(SignMethodHS256, accClaims).Sign(auth.key); err != nil {
				return t, err
			}
			if t.RefreshToken, err = NewToken(SignMethodHS256, refClaims).Sign(auth.key); err != nil {
				return t, err
			}
			return t, nil
		}
		return t, err
	}
	return DefaultToken{
		AccessToken:  accessToken,
		RefreshToken: refreshrToken,
	}, nil
}

// RevokeToken 撤销过期无效的Token
//revokedTokens     = make(map[string]time.Time) // 已撤销的令牌
//func init() {
//	ticker := time.NewTicker(60 * time.Second)
//	go func() {
//		for range ticker.C {
//			now := time.Now()
//			for token, expiresIn := range revokedTokens {
//				if now.After(expiresIn) {
//					delete(revokedTokens, token)
//				}
//			}
//		}
//	}()
//}
//
//// SetInvalidToken 将Token设置为无效
//func SetInvalidToken(token string) error {
//	claims := &Claims{}
//	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
//	if err != nil {
//		return err
//	}
//	if claims.ExpiresAt != nil {
//		revokedTokens[token] = claims.ExpiresAt.Time
//	}
//	return nil
//}
