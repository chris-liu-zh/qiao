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
	"context"
	"errors"
	"net/http"
	"time"
)

type timeoutWriter struct {
	w           http.ResponseWriter
	h           http.Header
	code        int
	closed      bool
	buf         bytes.Buffer
	wroteHeader bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	if tw.closed {
		return
	}
	tw.code = code
	tw.wroteHeader = true
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	if tw.closed {
		return 0, nil
	}
	return tw.buf.Write(p)
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }

func (router *RouterHandle) requestTimeout(w http.ResponseWriter, r *http.Request) {
	ctx := router.ctx
	if ctx == nil {
		var cancelCtx context.CancelFunc
		ctx, cancelCtx = context.WithTimeout(r.Context(), router.timeout)
		defer cancelCtx()
	}
	r = r.WithContext(ctx)
	done := make(chan struct{})
	hdr := make(http.Header)
	for k, v := range w.Header() {
		hdr[k] = append([]string(nil), v...)
	}

	tw := &timeoutWriter{
		w:    w,
		h:    hdr,
		code: http.StatusOK,
	}
	go func() {
		router.mux.ServeHTTP(tw, r)
		close(done)
	}()

	select {
	case <-done:
		buf := tw.buf.Bytes()
		code := tw.code
		wroteHeader := tw.wroteHeader
		hdr := tw.h

		for k, vals := range hdr {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}

		if wroteHeader {
			w.WriteHeader(code)
		}
		if len(buf) > 0 {
			w.Write(buf)
		}
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			TimeoutFail(w)
			tw.closed = true
		}
	}
}

func (router *RouterHandle) SetTimeout(timeout time.Duration) {
	router.timeout = timeout
}
