package qiao

import (
	"fmt"
	"testing"

	"github.com/chris-liu-zh/qiao/jwt"
)

type MyClaims struct {
	jwt.RegisteredClaims
	UserInfo any `json:"user_info"`
}

const (
	SecretKey = "123456"
)

func Test_Token(t *testing.T) {
	myClaims := MyClaims{
		UserInfo: "123",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "123",
		},
	}
	tokenstr, err := jwt.NewToken(jwt.SignMethodHS256, myClaims).Sign([]byte(SecretKey))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tokenstr)
	pmyClaims := &MyClaims{}
	token, err := jwt.ParseWithClaims(tokenstr, pmyClaims, []byte(SecretKey))
	if err != nil {
		t.Error(err)
		return
	}
	a, ok := token.Claims.(*MyClaims)
	if !ok {
		t.Error("user info error")
		return
	}
	fmt.Println(a.UserInfo)
}
