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
	Data    any    `json:"data"`
	Debug   error  `json:"debug,omitempty"`
	Success bool   `json:"success"`
}

func Success(data any) *Response {
	return &Response{Code: http.StatusOK, Message: "ok", Data: data, Success: true}
}

func Fail(message string, err error) *Response {
	if err == nil {
		return &Response{Code: http.StatusNotFound, Message: message}
	}
	return &Response{Code: http.StatusNotFound, Message: message, Debug: err}
}

func TimeoutFail() *Response {
	return &Response{Code: http.StatusRequestTimeout, Message: "Request timeout"}
}

func AuthFail() *Response {
	return &Response{Code: http.StatusUnauthorized, Message: "Unauthorized"}
}

func SignFail() *Response {
	return &Response{Code: http.StatusBadRequest, Message: "Sign fail"}
}

func TokenExpire() *Response {
	return &Response{Code: 498, Message: "Token expire"}
}

func (r *Response) SetCode(code int) *Response {
	r.Code = code
	return r
}

func (r *Response) SetMsg(msg string) *Response {
	r.Message = msg
	return r
}

func (r *Response) SetData(data any) *Response {
	r.Data = data
	return r
}

func (r *Response) WriteJson(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(r.Code)
	json.NewEncoder(w).Encode(r)
}

func SuccessJson(w http.ResponseWriter, data any) {
	Success(data).WriteJson(w)
}

func FailJson(w http.ResponseWriter, message string, err error) {
	Fail(message, err).WriteJson(w)
}

func (r *Response) Write(w http.ResponseWriter, data []byte) {
	w.WriteHeader(r.Code)
	w.Write(data)
}
