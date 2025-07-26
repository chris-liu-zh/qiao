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

type Claims struct {
	jwt.RegisteredClaims
}

func getJWTTime(t time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(t))
}

// createToken 创建Token
func (claims *Claims) CreateToken(key []byte) (string, error) {
	jt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jt.SignedString(key)
}

// VerifyToken 验证Token
func VerifyToken(token, issuer string, key []byte) (claims *Claims, err error) {
	claims = &Claims{}
	verifyToken, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return key, nil
	})
	if err != nil {
		return
	}
	if !verifyToken.Valid {
		return claims, errors.New("verify token failed")
	}
	if claims.Issuer != issuer {
		return nil, errors.New("issuer mismatch")
	}
	return claims, nil
}
