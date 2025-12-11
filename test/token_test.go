package qiao

import (
	"fmt"
	"testing"
	"time"

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
	if err := jwt.VerifyToken(tokenStr, &pmyClaims, []byte(SecretKey)); err != nil {
		t.Error(err)
		return
	}
	fmt.Println(pmyClaims.GetExpirationTime())
}

func Test_TokenAuth(t *testing.T) {

	jwt.SetAuth("api", ATExp, RTExp, "1D4JWUEGWWFK94JB74W1YGP9OF4L205F")
	data := Userdata{
		ID:   1,
		Name: "123",
	}

	token, err := jwt.CreateToken("api", jwt.WithUserInfo(data), jwt.WithSubject("123"))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(token.Token)
	fmt.Println(token.RefreshToken)

	time.Sleep(ATExp)
	//at := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJhcGkiLCJleHAiOiIyMDI1LTA4LTA3VDIzOjQwOjA4KzA4OjAwIiwidXNlcl9pbmZvIjp7InVpZCI6MSwidXNlcm5hbWUiOiIxMjMifX0.k7etCPO8ZdItPMK-_gkX0ooJGjfyEh770LCuhrcmDWk"
	token, err = jwt.RefreshToken("api", token.Token, token.RefreshToken)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(token.Token)
	fmt.Println(token.RefreshToken)
	getData := Userdata{}
	if err = jwt.CheckToken("api", token.Token, &getData); err != nil {
		t.Error(err)
		return
	}
	fmt.Println(getData)
}
