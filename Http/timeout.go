/*
 * @Author: Chris.Liu
 * @Date: 2024-07-31 11:55:13
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-10 11:37:29
 * @Description: 请填写简介
 */
package Http

import (
	"net/http"
	"sync"
)

type timeoutWriter struct {
	w http.ResponseWriter
	h http.Header
	// wbuf   bytes.Buffer
	// req    *http.Request
	code   int
	mu     sync.Mutex
	closed bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	if tw.closed {
		return
	}
	tw.code = code
	tw.w.WriteHeader(code)
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.closed {
		return 0, nil
	}
	return tw.w.Write(p)
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }
