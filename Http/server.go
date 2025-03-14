/*
 * @Author: Chris
 * @Date: 2023-06-13 23:17:20
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-09 22:26:01
 * @Description: 请填写简介
 */
package Http

import (
	"context"
	"net/http"
	"time"
)

type H struct {
	Addr   string
	Router *RouterHandle
	Server *http.Server
	Err    error
	Status bool
}

/**
 * @description:
 * @param {string} address
 * @param {*Router} router
 * @return {*}
 */
func NewHttpServer(address string, router *RouterHandle) *H {
	return &H{Addr: address, Router: router}
}

/**
 * @description:
 * @return {*}
 */
func (h *H) Start() {
	h.Server = &http.Server{Addr: h.Addr, Handler: h.Router}
	h.Status = true
	if err := h.Server.ListenAndServe(); err != nil {
		h.Status = false
		h.Err = err
	}
}

func (h *H) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if h.Status {
		if err := h.Server.Shutdown(ctx); nil != err {
			return err
		}
		return nil
	}
	return nil
}
