/*
 * @Author: Chris
 * @Date: 2025-03-09 16:24:53
 * @LastEditors: Strong
 * @LastEditTime: 2025-03-22 16:18:14
 * @Description: 请填写简介
 */
package qiao

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chris-liu-zh/qiao"

	"github.com/chris-liu-zh/qiao/Http"
)

const (
	ATExp = 1 * time.Hour
	RTExp = 72 * time.Hour
)

func Test_Http(t *testing.T) {
	Http.DefaultAuth("api", ATExp, RTExp, "1D4JWUEGWWFK94JB74W1YGP9OF4L205F")
	// if err := Http.NewTemplates("template/*.html", "template/**/*.html"); err != nil {
	// 	log.Println(err)
	// }

	if err := Http.NewLog("log", 10, 5, 30, true, true); err != nil {
		log.Println(err)
	}
	if err := qiao.NewLog().SetDefault(); err != nil {
		slog.Error("init logger error", "error", err)
	}

	r := Http.NewRouter()
	r.SetOnEvicted(onEvicted)
	r.SetTimeout(2 * time.Second)
	r.SetHeader(Http.DefaultHeader)
	r.SetContextSetter(setContest)
	r.SetSign("/api/", sign)
	r.SetAuth("/api/user/", auth)
	r.Get("/version", GetVersion)
	r.Get("/users/{id}", GetUserByID)
	r.Get("/", home)
	Http.NewHttpServer(":8080", r).Start()
}

func home(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Public string `json:"public"`
	}
	data.Public = "public"
	// Http.Success(data).Json(w)
	json.NewEncoder(w).Encode(data)
	// Http.Html(w, "template/index", data)
}

func onEvicted(w http.ResponseWriter, r *http.Request) {
	h := Http.GetHeader(r)
	if r.URL.Path == "/logout" {
		if token, ok := h["Authorization"]; ok {
			Http.SetInvalidToken(token)
		}
	}
}

func setContest(r *http.Request) *http.Request {
	return Http.SetContext(r, "user", "admin")
}

func GetUserByID(w http.ResponseWriter, r *http.Request) {
	user, ok := Http.GetContext(r, "user").(string)
	if ok {
		log.Printf("获取到的用户值: %s", user)
	} else {
		log.Println("未获取到用户值")
	}

	id := r.PathValue("id")
	response := map[string]string{
		"id":   id,
		"name": user,
	}
	Http.SuccessJson(w, response)
}

func GetVersion(w http.ResponseWriter, r *http.Request) {
	var ver struct {
		Version string `json:"version"`
		Ip      string `json:"ip"`
	}
	ver.Version = "1.0.0"
	ver.Ip = strings.Split(r.RemoteAddr, ":")[0]
	time.Sleep(3 * time.Second)
	w.Header().Set("content-type", "application/json;charset=UTF-8")
	Http.SuccessJson(w, ver)
	// Http.Html(w, "template/version/index", ver)
}

func sign(header map[string]string) error {
	const (
		key    = "ALYDDNS"
		secret = "1D4JWUEGWWFK94JB74W1YGP9OF4L205F"
	)
	sign := strings.ToUpper(header["Sign"])
	timestampStr := header["Timestamp"]

	// 将时间戳字符串转换为时间类型
	timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return errors.New("timestamp 解析错误")
	}
	// 使用 time.Unix 函数将时间戳转换为 time.Time 类型
	timestamp := time.Unix(timestampInt, 0)

	return Http.DefaultSign(sign, key, secret, timestamp, 5*time.Minute)
}

func auth(header map[string]string) (contextKey Http.CtxKey, data any, err error) {
	token, ok := header["Authorization"]
	if !ok {
		return "", nil, errors.New("token not found")
	}
	if data, err = Http.CheckToken("api", token); err != nil {
		return "", nil, err
	}
	return "userinfo", data, nil
}
