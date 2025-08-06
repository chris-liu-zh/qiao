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

type Userdata struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

const (
	SecretKey = "123456"
)

func newClaims(data any) MyClaims {
	return MyClaims{
		UserInfo: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "123",
		},
	}
}

func Test_Token(t *testing.T) {
	data := Userdata{
		ID:   1,
		Name: "123",
	}
	myClaims := newClaims(data)
	tokenStr, err := jwt.NewToken(jwt.SignMethodHS256, myClaims).Sign([]byte(SecretKey))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tokenStr)

	pmyClaims := newClaims(&Userdata{})
	fmt.Println(pmyClaims.UserInfo)
	if err := jwt.VerifyToken(tokenStr, &pmyClaims, []byte(SecretKey)); err != nil {
		t.Error(err)
		return
	}
	fmt.Println(pmyClaims.UserInfo.(*Userdata).Name)
}
