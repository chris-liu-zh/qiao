/*
 * @Author: Chris
 * @Date: 2024-05-31 14:03:31
 * @LastEditors: Strong
 * @LastEditTime: 2025-03-21 15:55:30
 * @Description: 请填写简介
 */
package Http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/chris-liu-zh/qiao/jwt"
)

type RouterHandle struct {
	m         middleware
	Next      http.Handler
	timeout   time.Duration
	ctx       context.Context
	mux       *http.ServeMux
	onEvicted func(http.ResponseWriter, *http.Request)
}

type CtxKey string
type logResponseWriter struct {
	http.ResponseWriter
	status       int //用于记录响应状态码
	bytesWritten int // 用于记录响应字节数
	msg          string
}

func (lw *logResponseWriter) WriteHeader(code int) {
	lw.status = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *logResponseWriter) Write(b []byte) (int, error) {
	n, err := lw.ResponseWriter.Write(b)
	lw.bytesWritten += n // 记录写入的字节数
	lw.msg = string(b)
	return n, err
}

func GetHeader(r *http.Request) map[string]string {
	h := make(map[string]string)
	for name, values := range r.Header {
		for _, value := range values {
			h[name] = value
		}
	}
	return h
}

func (router *RouterHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if router.Next == nil {
		router.Next = http.DefaultServeMux
	}
	if router.onEvicted != nil {
		router.onEvicted(w, r)
	}

	lw := &logResponseWriter{ResponseWriter: w}
	if router.m.setHeader != nil {
		router.m.setHeader(lw)
	}
	if r.Method == "OPTIONS" {
		return
	}
	if _, pattern := router.mux.Handler(r); pattern == "" {
		Fail("404 Not Found", nil).WriteJson(lw)
		return
	}
	header := GetHeader(r)
	if err := router.m.sign(r.URL.Path, header); err != nil {
		SignFail().WriteJson(lw)
		LogError(r, lw.status, lw.bytesWritten, lw.msg)
		return
	}

	key, userinfo, err := router.m.auth(r.URL.Path, header)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			TokenExpire().WriteJson(lw)
			LogError(r, lw.status, lw.bytesWritten, lw.msg)
			return
		}
		AuthFail().WriteJson(lw)
		LogError(r, lw.status, lw.bytesWritten, lw.msg)
		return
	}

	if userinfo != nil && key != "" {
		// 将用户信息存储在请求上下文中
		ctx := context.WithValue(r.Context(), key, userinfo)
		r = r.WithContext(ctx)
	}

	if router.m.contextSetter != nil {
		r = router.m.contextSetter(r)
	}

	if router.timeout > 0 {
		router.requestTimeout(lw, r)
	} else {
		router.mux.ServeHTTP(lw, r)
	}

	if lw.status >= http.StatusBadRequest {
		LogError(r, lw.status, lw.bytesWritten, lw.msg)
	} else {
		LogAccess(r, lw.status, lw.bytesWritten)
	}
}

func (router *RouterHandle) Get(path string, handler http.HandlerFunc) {
	router.mux.HandleFunc(fmt.Sprintf("GET %s", path), handler)
}

func (router *RouterHandle) Post(path string, handler http.HandlerFunc) {
	router.mux.HandleFunc(fmt.Sprintf("POST %s", path), handler)
}

// 文件服务路由
func (router *RouterHandle) FilesServer(path, dir string) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Println(err)
	}

	fs := http.StripPrefix(path, http.FileServer(http.Dir(dir)))
	router.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

func (router *RouterHandle) SetOnEvicted(f func(http.ResponseWriter, *http.Request)) {
	router.onEvicted = f
}

func SetContext(r *http.Request, key string, value any) *http.Request {
	ctx := context.WithValue(r.Context(), CtxKey(key), value)
	return r.WithContext(ctx)
}

func GetContext(r *http.Request, key string) any {
	return r.Context().Value(CtxKey(key))
}

func NewRouter() *RouterHandle {
	return &RouterHandle{
		mux: http.NewServeMux(),
		m: middleware{
			Auth: make(map[string]auth),
			Sign: make(map[string]sign),
		},
	}
}
