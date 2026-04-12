// Copyright 2013 Julien Schmidt. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE.

package rdx

import (
	"net/http"

	"github.com/suisrc/zoo"
)

// -----------------------------------------------------------------------------------
// 基于 Tire / Radix Tree 的 httprouter 的路由
// 也是为外部扩展路由提供标准实现方式
// 路由算法引用 https://github.com/julienschmidt/httprouter

func init() {
	zoo.Engines["rdx"] = NewRdxRouter
}

// var _ zoo.Engine = (*RdxRouter)(nil)

func NewRdxRouter(svckit zoo.SvcKit) zoo.Engine {
	return &RdxRouter{
		name:   "zoo-rdx",
		svckit: svckit,
		Router: New(),
	}
}

type RdxRouter struct {
	name   string
	svckit zoo.SvcKit
	Router *Router
}

func (aa *RdxRouter) Name() string {
	return aa.name
}

func (aa *RdxRouter) Handle(method, action string, handle zoo.HandleFunc) {
	path := "/" + action
	if method == "" {
		method = "GET" // 默认使用 GET
	}
	aa.Router.Handle(method, path, func(rw http.ResponseWriter, rr *http.Request, ps Params) {
		ctx := zoo.NewCtx(aa.svckit, rr, rw, aa.name)
		ctx.Params = ps.ByName
		defer ctx.Clear()
		handle(ctx)
	})
}

func (aa *RdxRouter) ServeHTTP(rw http.ResponseWriter, rr *http.Request) {
	aa.Router.ServeHTTP(rw, rr)
}
