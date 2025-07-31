/*
 * @Author: Chris
 * @Date: 2024-06-12 13:24:32
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-14 16:53:09
 * @Description: 请填写简介
 */
package Http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

type auth func(map[string]string) (CtxKey, Userinfo, error)
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

func (router *RouterHandle) requestTimeout(w http.ResponseWriter, r *http.Request) {
	ctx := router.ctx
	if ctx == nil {
		var cancelCtx context.CancelFunc
		ctx, cancelCtx = context.WithTimeout(r.Context(), router.timeout)
		defer cancelCtx()
	}
	r = r.WithContext(ctx)
	done := make(chan struct{})
	tw := &timeoutWriter{
		w:    w,
		h:    w.Header(),
		req:  r,
		code: http.StatusOK,
	}
	go func() {
		router.mux.ServeHTTP(tw, r)
		close(done)
	}()

	select {
	case <-done:
		//log.Println("request completed")
	case <-ctx.Done():
		switch err := ctx.Err(); {
		case errors.Is(err, context.DeadlineExceeded):
			TimeoutFail().WriteJson(tw)
			tw.closed = true
		default:
			tw.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}

func (router *RouterHandle) SetTimeout(timeout time.Duration) {
	router.timeout = timeout
}
