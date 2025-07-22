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
	"net/http"
)

type Return struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Debug   error  `json:"debug"`
	Success bool   `json:"success"`
}

func Success(data any) *Return {
	return &Return{Code: http.StatusOK, Message: "ok", Data: data, Success: true}
}

func LoginFail(message string, err error) *Return {
	return &Return{Code: http.StatusUnauthorized, Message: message, Debug: err}
}

func Fail(message string, err error) *Return {
	if err == nil {
		return &Return{Code: http.StatusNotFound, Message: message}
	}
	return &Return{Code: http.StatusNotFound, Message: message, Debug: err}
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
	return &Return{Code: 498, Message: "Token expire"}
}

func (r *Return) Json(w http.ResponseWriter) {
	if r.Code != 200 {
		LogDebug(r.Message, 2)
	}
	w.Header().Set("content-type", "application/json;charset=UTF-8")
	dataByte, _ := json.Marshal(r)
	HttpWrite(w, dataByte, r.Code, r.Message)
}

func HttpWrite(w http.ResponseWriter, data []byte, code int, msg string) {
	w.WriteHeader(code)
	w.Write(data)
}
