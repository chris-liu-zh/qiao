/*
 * @Author: Chris
 * @Date: 2022-05-29 16:32:22
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-11 13:42:34
 * @Description: 请填写简介
 */
package Http

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrorInvalidToken   = errors.New("verify Token Failed")
	ErrorTokenExpire    = errors.New("token has expired")
	ErrorTokenNotExpire = errors.New("token not expired")
)

type claims struct {
	UserInfo any `json:"user_info"`
	jwt.RegisteredClaims
}

var revokedTokens = make(map[string]time.Time)

type Token struct {
	accessTokenExpired  time.Duration
	refreshTokenExpired time.Duration
	secretKey           []byte
}

func NewToken(accessExpired, refreshExpired time.Duration, key string) *Token {
	CleanupInvalidToken(60 * time.Second)
	return &Token{
		accessTokenExpired:  accessExpired,
		refreshTokenExpired: refreshExpired,
		secretKey:           []byte(key),
	}
}

func getJWTTime(t time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(t))
}

func (t *Token) keyFunc(token *jwt.Token) (any, error) {
	return t.secretKey, nil
}

// CreateToken 颁发token access token 和 refresh token
func (t *Token) CreateToken(UserInfo any) (accessToken, refreshToken string, err error) {
	accessClaims := claims{
		UserInfo,
		jwt.RegisteredClaims{
			ExpiresAt: getJWTTime(t.accessTokenExpired),
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(t.secretKey)
	if err != nil {
		return
	}
	refreshClaims := claims{
		UserInfo,
		jwt.RegisteredClaims{
			ExpiresAt: getJWTTime(t.refreshTokenExpired),
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(t.secretKey)
	if err != nil {
		return
	}
	return
}

func (t *Token) SetInvalidToken(token string) error {
	claims := &claims{}
	if _, err := jwt.ParseWithClaims(token, claims, t.keyFunc); err != nil {
		return err
	}
	if claims.ExpiresAt != nil {
		revokedTokens[token] = claims.ExpiresAt.Time
	}
	return nil
}

func CleanupInvalidToken(interval time.Duration) {
	ticker := time.NewTicker(interval)
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

// VerifyToken 验证Token
func (t *Token) VerifyToken(accessToken string) (*claims, error) {
	if _, ok := revokedTokens[accessToken]; ok {
		return nil, ErrorInvalidToken
	}
	accessClaims := &claims{}
	verifyToken, err := jwt.ParseWithClaims(accessToken, accessClaims, t.keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrorTokenExpire
		}
		return nil, err
	}
	if !verifyToken.Valid {
		return nil, ErrorInvalidToken
	}
	return accessClaims, nil
}

// RefreshToken 通过 refresh token 刷新 atoken
func (t *Token) RefreshToken(accessToken, refreshToken string) (at string, rt string, err error) {
	if _, ok := revokedTokens[accessToken]; ok {
		return "", "", ErrorInvalidToken
	}
	if _, ok := revokedTokens[refreshToken]; ok {
		return "", "", ErrorInvalidToken
	}
	// refresh Token 无效直接返回
	if _, err = jwt.Parse(refreshToken, t.keyFunc); err != nil {
		return
	}
	// 从旧access token 中解析出claims数据
	accessClaims := &claims{}
	if _, err = jwt.ParseWithClaims(accessToken, accessClaims, t.keyFunc); err != nil {
		// 判断错误是不是因为access token 正常过期导致的
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return t.CreateToken(accessClaims.UserInfo)
		}
		return
	}
	return "", "", ErrorTokenNotExpire
}
