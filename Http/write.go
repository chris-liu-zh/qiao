/*
 * @Author: Chris
 * @Date: 2023-03-29 11:04:12
 * @LastEditors: Strong
 * @LastEditTime: 2025-03-21 15:54:08
 * @Description: 请填写简介
 */
package Http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Return struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
}

func Success(data any) *Return {
	return &Return{Code: http.StatusOK, Message: "ok", Data: data, Success: true}
}

func Fail(err error, message ...string) *Return {
	return &Return{Code: http.StatusNotFound, Message: fmt.Sprintf("%v:%s", err, message)}
}

func TimeoutFail() *Return {
	return &Return{Code: http.StatusRequestTimeout, Message: "Request timeout"}
}

func AuthFail() *Return {
	return &Return{Code: http.StatusUnauthorized, Message: "Unauthorized"}
}

func SignFail() *Return {
	return &Return{Code: http.StatusBadRequest, Message: "Sign fail"}
}

func TokenExpire() *Return {
	return &Return{Code: http.StatusForbidden, Message: "Token expire"}
}

// func debug(msg string, skip int) {
// 	if msg != "" {
// 		if funcName, file, line, ok := runtime.Caller(skip); ok {
// 			log.Println(msg, file, line, runtime.FuncForPC(funcName).Name())
// 		}
// 	}
// }

func (r *Return) Json(w http.ResponseWriter) {
	dataByte, err := json.Marshal(r)
	if err != nil {
		Fail(err).Json(w)
		return
	}
	// if r.Code != 200 {
	// 	HttpWrite(w, dataByte, r.Code, r.Msg)
	// 	return
	// }
	w.Header().Set("content-type", "application/json;charset=UTF-8")
	HttpWrite(w, dataByte, r.Code, r.Message)
}

func HttpWrite(w http.ResponseWriter, data []byte, code int, msg string) {
	w.WriteHeader(code)
	w.Write(data)
}
