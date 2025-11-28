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

type RESTful struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data,omitempty"`
	Debug   any    `json:"debug,omitempty"`
	Success bool   `json:"success"`
}

func Success(w http.ResponseWriter, data any, debug ...any) {
	WriteJson(w, http.StatusOK, "ok", data, debug)
}

func Error(w http.ResponseWriter, code int, message string, debug ...any) {
	WriteJson(w, code, message, nil, debug)
}

func BadRequest(w http.ResponseWriter, msg string, debug ...any) {
	WriteJson(w, http.StatusBadRequest, msg, nil, debug)
}

func Forbidden(w http.ResponseWriter, msg string, debug ...any) {
	WriteJson(w, http.StatusForbidden, msg, nil, debug)
}

func NotFound(w http.ResponseWriter, msg string, debug ...any) {
	WriteJson(w, http.StatusNotFound, msg, nil, debug)
}

func TimeoutFail(w http.ResponseWriter, debug ...any) {
	WriteJson(w, http.StatusRequestTimeout, "Request timeout", nil, debug)
}

func Unauthorized(w http.ResponseWriter, msg string, debug ...any) {
	WriteJson(w, http.StatusUnauthorized, msg, nil, debug)
}

func WriteJson(w http.ResponseWriter, code int, message string, data any, debug ...any) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(code)
	r := &RESTful{Code: code, Message: message, Data: data, Debug: debug}
	if code <= 400 {
		r.Success = true
	}
	json.NewEncoder(w).Encode(r)
}
