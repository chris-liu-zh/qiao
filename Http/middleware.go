/*
 * @Author: Chris
 * @Date: 2024-06-12 13:24:32
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-14 16:53:09
 * @Description: 请填写简介
 */
package Http

import (
	"net/http"
	"strings"
)

type auth func(map[string]string) (CtxKey, any, error)
type sign func(map[string]string) error

type middleware struct {
	Auth          map[string]auth
	Sign          map[string]sign
	setHeader     func(w http.ResponseWriter)
	contextSetter func(*http.Request) *http.Request
}

func (router *RouterHandle) SetHeader(f func(w http.ResponseWriter)) {
	router.m.setHeader = f
}

func (router *RouterHandle) SetContextSetter(setter func(*http.Request) *http.Request) {
	router.m.contextSetter = setter
}

func (router *RouterHandle) SetAuth(authPath string, authFunc auth) {
	router.m.Auth[authPath] = authFunc
}

func (router *RouterHandle) SetSign(signPath string, signfunc sign) {
	router.m.Sign[signPath] = signfunc
}

func (m *middleware) sign(url string, header map[string]string) (err error) {
	for p, f := range m.Sign {
		if strings.HasPrefix(url, p) {
			if err = f(header); err != nil {
				return
			}
		}
	}
	return
}

func (m *middleware) auth(url string, header map[string]string) (userInfoKey CtxKey, authInfo any, err error) {
	for p, f := range m.Auth {
		if strings.HasPrefix(url, p) {
			return f(header)
		}
	}
	return
}

func Permission(authName string, authorityFunc func(*http.Request, string) bool, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorityFunc(r, authName) {
			Forbidden(w, "Forbidden")
			return
		}
		handler(w, r)
	}
}
