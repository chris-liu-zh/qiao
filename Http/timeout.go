/*
 * @Author: Chris.Liu
 * @Date: 2024-07-31 11:55:13
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-10 11:37:29
 * @Description: 请填写简介
 */
package Http

import (
	"bytes"
	"net/http"
	"sync"
)

type timeoutWriter struct {
	w    http.ResponseWriter
	h    http.Header
	wbuf bytes.Buffer
	req  *http.Request
	code int
	mu   sync.Mutex
	err  error
}

func (lw *timeoutWriter) WriteHeader(code int) {
	lw.code = code
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.err != nil {
		return 0, tw.err
	}
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }
