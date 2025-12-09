/*
 * @Author: Chris.Liu
 * @Date: 2024-07-31 11:55:13
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-10 11:37:29
 * @Description: 请填写简介
 */
package Http

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
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

func (router *RouterHandle) requestTimeout(w http.ResponseWriter, r *http.Request) {
	ctx := router.ctx
	if ctx == nil {
		var cancelCtx context.CancelFunc
		ctx, cancelCtx = context.WithTimeout(r.Context(), router.timeout)
		defer cancelCtx()
	}
	// r = r.WithContext(ctx)
	done := make(chan struct{})
	tw := &timeoutWriter{
		w: w,
		h: w.Header(),
		// req:  r,
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
			TimeoutFail(tw)
			tw.closed = true
		default:
			tw.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}

func (router *RouterHandle) SetTimeout(timeout time.Duration) {
	router.timeout = timeout
}
