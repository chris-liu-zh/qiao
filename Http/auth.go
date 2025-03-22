package Http

/*
 * @Author: Chris
 * @Date: 2023-06-13 14:17:57
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-22 12:35:52
 * @Description: 请填写简介
 */

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/chris-liu-zh/qiao"
	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	key           []byte
	accessClaims  jwt.RegisteredClaims
	refreshClaims jwt.RegisteredClaims
}

var revokedTokens = make(map[string]time.Time)

func DefaultAuth(issuer string, aExp, rExp time.Duration, key string) *Auth {
	return &Auth{
		key:           []byte(key),
		accessClaims:  CreateClaims(issuer, aExp),
		refreshClaims: CreateClaims(issuer, rExp),
	}
}

func NewAuth(issuer string, aExp time.Duration, key string) *Auth {
	return &Auth{
		key:          []byte(key),
		accessClaims: CreateClaims(issuer, aExp),
	}
}

/**
 * @description:验证签名
 * @param {*} appkey
 * @param {*} sign
 * @param {string} timestamp
 * @return {*}
 */
func DefaultSign(header map[string]string, appKey, secret string, timeDiff time.Duration) error {
	sign := strings.ToUpper(header["Sign"])
	timestampStr := header["Timestamp"]

	// 将时间戳字符串转换为时间类型
	timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return errors.New("timestamp 解析错误")
	}
	// 使用 time.Unix 函数将时间戳转换为 time.Time 类型
	timestamp := time.Unix(timestampInt, 0)

	now := time.Now()

	// 检查时间戳是否在有效时间范围内
	if timestamp.Before(now.Add(-timeDiff)) || timestamp.After(now.Add(timeDiff)) {
		return errors.New("timestamp 超出有效时间范围，请检查系统时间")
	}

	localSign := strings.ToUpper(qiao.MD5(appKey + timestampStr + secret))

	if localSign != sign {
		return errors.New("sign error")
	}
	return nil
}

// 创建 JWT 注册声明
func CreateClaims(issuer string, exp time.Duration) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    issuer,
		ExpiresAt: getJWTTime(exp),
	}
}

/**
 * @description: 获取 access token 和 refresh token
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func (a *Auth) NewDefaultToken(data any) (aToken, rToken string, err error) {
	aToken, err = CreateToken(data, a.accessClaims, a.key)
	if err != nil {
		return
	}
	rToken, err = CreateToken(nil, a.refreshClaims, a.key)
	if err != nil {
		return
	}
	return
}

/**
 * @description: 获取 Token
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func (a *Auth) NewToken(data any) (string, error) {
	return CreateToken(data, a.accessClaims, a.key)
}

/**
 * @description: 验证Token
 * @param {*http.Request} r
 * @return {*}
 */
func (a *Auth) CheckToken(token string) (any, error) {
	claims, err := VerifyToken(token, a.key)
	if err != nil {
		return nil, err
	}
	return claims.UserInfo, nil
}

// 刷新Token
func (a *Auth) RefreshToken(accessToken, refreshToken string) (string, string, error) {
	if _, err := a.CheckToken(refreshToken); err != nil {
		return "", "", err
	}

	if userinfo, err := a.CheckToken(accessToken); err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return a.NewDefaultToken(userinfo)
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
	claims := &claims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)
	if err != nil {
		return err
	}
	if claims.ExpiresAt != nil {
		revokedTokens[token] = claims.ExpiresAt.Time
	}
	return nil
}
