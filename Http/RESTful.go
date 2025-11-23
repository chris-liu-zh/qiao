/*
		package Http
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

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data,omitempty"`
	Success bool   `json:"success"`
}

func Success(w http.ResponseWriter, data ...any) {
	WriteJson(w, http.StatusOK, "ok", data)
}

func BadRequest(w http.ResponseWriter, msg string) {
	WriteJson(w, http.StatusBadRequest, msg, nil)
}

func Forbidden(w http.ResponseWriter, msg string) {
	WriteJson(w, http.StatusForbidden, msg, nil)
}

func NotFound(w http.ResponseWriter, msg string) {
	WriteJson(w, http.StatusNotFound, msg, nil)
}

func TimeoutFail(w http.ResponseWriter) {
	WriteJson(w, http.StatusRequestTimeout, "Request timeout", nil)
}

func Unauthorized(w http.ResponseWriter, msg string) {
	WriteJson(w, http.StatusUnauthorized, msg, nil)
}

func WriteJson(w http.ResponseWriter, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(code)
	r := &Response{Code: code, Message: message, Data: data}
	if code == http.StatusOK {
		r.Success = true
	}
	json.NewEncoder(w).Encode(r)
}
