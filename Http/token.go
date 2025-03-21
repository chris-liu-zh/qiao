/*
 * @Author: Chris
 * @Date: 2022-05-29 16:32:22
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-21 13:12:02
 * @Description: 请填写简介
 */
package Http

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type claims struct {
	UserInfo any `json:"user_info"`
	jwt.RegisteredClaims
}

func getJWTTime(t time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(t))
}

func CreateToken(UserInfo any, reg jwt.RegisteredClaims, key []byte) (string, error) {
	tokenClaims := claims{
		UserInfo,
		reg,
	}
	jt := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return jt.SignedString(key)
}

// VerifyToken 验证Token
func VerifyToken(token string, key []byte) (*claims, error) {
	claims := &claims{}
	verifyToken, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if !verifyToken.Valid {
		return nil, errors.New("verify token failed")
	}
	return claims, nil
}
