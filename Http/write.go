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
	Message string `json:"msg"`
	Data    any    `json:"data"`
	Debug   error  `json:"debug"`
	Success bool   `json:"success"`
}

func Success(data any) *Return {
	return &Return{Code: http.StatusOK, Message: "ok", Data: data, Success: true}
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

func (r *Return) SetCode(code int) *Return {
	r.Code = code
	return r
}

func (r *Return) Json(w http.ResponseWriter) {
	if r.Code != 200 {
		LogDebug(r.Message, 2)
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	dataByte, _ := json.Marshal(r)
	Write(w, dataByte, r.Code)
}

func SuccessJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	json.NewEncoder(w).Encode(Success(data))
}

func Write(w http.ResponseWriter, data []byte, code int) {
	w.WriteHeader(code)
	w.Write(data)
}
