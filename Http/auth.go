package Http

/*
 * @Author: Chris
 * @Date: 2023-06-13 14:17:57
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-14 12:09:22
 * @Description: 请填写简介
 */

import (
	"errors"
	"strings"
	"time"

	"github.com/chris-liu-zh/qiao"
)

type Auth struct {
	token   string
	refresh string
	t       *Token
}

func DefaultAuth(AccessTokenExp, RefreshTokenExp time.Duration, secretKey string) *Auth {
	return &Auth{
		t: NewToken(AccessTokenExp, RefreshTokenExp, secretKey),
	}
}

func (a *Auth) SetInvalidToken(token string) error {
	return a.t.SetInvalidToken(token)
}

/**
 * @description:验证签名
 * @param {*} appkey
 * @param {*} sign
 * @param {string} timestamp
 * @return {*}
 */
func DefaultSign(header map[string]string, key, secret string) error {
	appkey := header["Appkey"]
	sign := strings.ToUpper(header["Sign"])
	timestamp := header["Timestamp"]

	localSign := strings.ToUpper(qiao.MD5(timestamp + secret))

	if appkey != key || localSign != sign {
		return errors.New("sign error")
	}
	return nil
}

/**
 * @description: 获取 Token
 * @param {http.ResponseWriter} w
 * @param {*http.Request} r
 * @return {*}
 */
func (a *Auth) GetToken(u any) (string, string, error) {
	if a == nil {
		return "", "", errors.New("非法访问")
	}
	return a.t.CreateToken(u)
}

/**
 * @description: 验证Token
 * @param {*http.Request} r
 * @return {*}
 */
func (a *Auth) CheckToken(header map[string]string) (any, error) {
	a.token = header["Authorization"]
	if a.token == "" {
		return nil, errors.New("拒绝访问")
	}
	claims, err := a.t.VerifyToken(a.token)
	if err != nil {
		return nil, err
	}
	return claims.UserInfo, nil
}

func (a *Auth) RefreshToken(header map[string]string) (string, string, error) {
	a.token = header["Authorization"]
	a.refresh = header["Refresh-Token"]
	if a.refresh == "" || a.token == "" {
		return "", "", errors.New("拒绝访问")
	}
	return a.t.RefreshToken(a.token, a.refresh)
}
