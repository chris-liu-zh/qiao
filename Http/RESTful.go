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
	Code        int    `json:"code"`
	Message     string `json:"msg"`
	Data        any    `json:"data,omitempty"`
	Debug       any    `json:"debug,omitempty"`
	Success     bool   `json:"success"`
	writeHeader bool   `json:"-"`
}

type options func(*RESTful)

func SetSuccess(success bool) options {
	return func(r *RESTful) {
		r.Success = success
	}
}

func SetDebug(debug any) options {
	return func(r *RESTful) {
		r.Debug = debug
	}
}

func Success(w http.ResponseWriter, data any, opt ...options) {
	WriteJson(w, http.StatusOK, "ok", data, opt...)
}

func SuccessNoContent(w http.ResponseWriter, opt ...options) {
	WriteJson(w, http.StatusNoContent, "no content", nil, opt...)
}

func SuccessCreated(w http.ResponseWriter, opt ...options) {
	WriteJson(w, http.StatusCreated, "created", nil, opt...)
}

func Error(w http.ResponseWriter, code int, message string, opt ...options) {
	WriteJson(w, code, message, nil, opt...)
}

func BadRequest(w http.ResponseWriter, msg string, opt ...options) {
	WriteJson(w, http.StatusBadRequest, msg, nil, opt...)
}

func Forbidden(w http.ResponseWriter, msg string, opt ...options) {
	WriteJson(w, http.StatusForbidden, msg, nil, opt...)
}

func NotFound(w http.ResponseWriter, msg string, opt ...options) {
	WriteJson(w, http.StatusNotFound, msg, nil, opt...)
}

func TimeoutFail(w http.ResponseWriter, opt ...options) {
	WriteJson(w, http.StatusRequestTimeout, "Request timeout", nil, opt...)
}

func Unauthorized(w http.ResponseWriter, msg string, opt ...options) {
	WriteJson(w, http.StatusUnauthorized, msg, nil, opt...)
}

func WriteJson(w http.ResponseWriter, code int, message string, data any, opt ...options) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	r := &RESTful{Code: code, Message: message, Data: data, writeHeader: true}
	if code <= 400 {
		r.Success = true
	}
	for _, o := range opt {
		o(r)
	}
	if r.writeHeader {
		w.WriteHeader(code)
	}
	json.NewEncoder(w).Encode(r)
}
