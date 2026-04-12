// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

package zoo

import (
	"net/http"
	"strings"
)

// engine: 引擎管理工具

// -----------------------------------------------------------------------------------

// 基于 map 路由，为更高的性能，单接口而生，是默认的路由
// var _ Engine = (*MapRouter)(nil)

func NewMapRouter(svckit SvcKit) Engine {
	return &MapRouter{
		name:    "zoo-map",
		svckit:  svckit,
		Handles: make(map[string]HandleFunc),
	}
}

type MapRouter struct {
	name    string
	svckit  SvcKit
	Handle_ HandleFunc            // 默认函数，没有找到Action触发
	Handles map[string]HandleFunc // 接口集合

	// https://github.com/puzpuzpuz/xsync
	// 初始化后，map 就不会变更了，可以使用 xsync.Map 获取更高的性能
	// handles *xsync.Map[string, HandleFunc]
}

func (aa *MapRouter) Name() string {
	return aa.name
}

func (aa *MapRouter) Handle(method, action string, handle HandleFunc) {
	if method == "" && action == "" {
		aa.Handle_ = handle // 默认函数
	} else {
		if method == "" {
			method = "GET" // 默认使用 GET
		}
		aa.Handles[method+" /"+action] = handle
	}
}

func (aa *MapRouter) GetHandle(method, action string) (HandleFunc, bool) {
	handle, exist := aa.Handles[method+" /"+action]
	return handle, exist
}

func (aa *MapRouter) ServeHTTP(rw http.ResponseWriter, rr *http.Request) {
	// 查询并执行业务 Action
	ctx := NewCtx(aa.svckit, rr, rw, aa.name)
	defer ctx.Clear() // 确保取消
	if ctx.Action == "healthz" {
		// 健康健康高优先级， 直接出发检索
		Healthz(ctx)
	} else if handle, exist := aa.GetHandle(rr.Method, ctx.Action); exist {
		// 处理函数
		handle(ctx)
	} else if aa.Handle_ != nil {
		// 默认函数
		aa.Handle_(ctx)
	} else if ctx.Action == "" {
		// 空的操作
		res := &Result{ErrCode: "action-empty", Message: "未指定操作: empty"}
		ctx.JSON(res)
	} else {
		// 无效操作
		res := &Result{ErrCode: "action-unknow", Message: "未指定操作: " + ctx.Action}
		ctx.JSON(res) // 无效操作
	}
}

// -----------------------------------------------------------------------------------

// 基于 http.ServeMux 的路由
// var _ Engine = (*MuxRouter)(nil)

func NewMuxRouter(svckit SvcKit) Engine {
	return &MuxRouter{
		name:   "zoo-mux",
		svckit: svckit,
		Router: http.NewServeMux(),
	}
}

type MuxRouter struct {
	name   string
	svckit SvcKit
	Router *http.ServeMux
}

func (aa *MuxRouter) Name() string {
	return aa.name
}

func (aa *MuxRouter) Handle(method, action string, handle HandleFunc) {
	pattern := "/" + action
	if method != "" {
		pattern = method + " " + pattern
	}
	aa.Router.HandleFunc(pattern, func(rw http.ResponseWriter, rr *http.Request) {
		ctx := NewCtx(aa.svckit, rr, rw, aa.name)
		defer ctx.Clear()
		handle(ctx)
	})
}

func (aa *MuxRouter) ServeHTTP(rw http.ResponseWriter, rr *http.Request) {
	aa.Router.ServeHTTP(rw, rr)
}

// -----------------------------------------------------------------------------------
// -----------------------------------------------------------------------------------

type RdeHelper interface {
	KeyGetter(key string) (string, error)
	NewRouter(svckit SvcKit) Engine
}

// -----------------------------------------------------------------------------------

// dir 识别 action，让后在进行 rdx 路由识别， 路由样式 {dir}/...

// var _ Engine = (*DirRouter)(nil)

func NewDirRouter(svckit SvcKit) Engine {
	helper, _ := svckit.Get("rde-helper").(RdeHelper)
	return &DirRouter{
		name:   "zoo-dir",
		svckit: svckit,
		Helper: helper,
		Router: make(map[string]Engine),
	}
}

type DirRouter struct {
	name   string
	svckit SvcKit
	Helper RdeHelper
	Router map[string]Engine
}

func (aa *DirRouter) Name() string {
	return aa.name
}

func (aa *DirRouter) Handle(method, action string, handle HandleFunc) {
	if aa.Helper == nil {
		// 如果 Helper 为空，直接返回，不注册任何路由
		return
	}
	if G.Server.ApiRoot != "" {
		action = strings.TrimPrefix("/"+action, G.Server.ApiRoot+"/")
	}
	// 分拆 action 为 key 和 path
	key, path := "", ""
	if idx := strings.Index(action, "/"); idx > 0 {
		key, path = action[:idx], action[idx+1:]
	} else {
		key, path = action, ""
	}
	// 获取路由是否存在，不存在则创建一个新的路由并注册到 aa.Router 中，最后在路由上注册 handle
	router, exist := aa.Router[key]
	if !exist {
		router = aa.Helper.NewRouter(aa.svckit)
		aa.Router[key] = router
	}
	router.Handle(method, path, handle)
}

func (aa *DirRouter) ServeHTTP(rw http.ResponseWriter, rr *http.Request) {
	if aa.Helper == nil {
		// 如果 Helper 为空，直接返回 404 Not Found
		http.NotFound(rw, rr)
		return
	}
	if G.Server.ApiRoot != "" {
		rr.URL.Path = strings.TrimPrefix(rr.URL.Path, G.Server.ApiRoot+"/")
		rr.URL.RawPath = strings.TrimPrefix(rr.URL.RawPath, G.Server.ApiRoot+"/")
	}
	// 从 URL.Path 中提取第一个目录作为 key
	action := rr.URL.Path
	if len(action) > 0 && action[0] == '/' {
		action = action[1:]
	}
	if idx := strings.Index(action, "/"); idx > 0 {
		key, err := aa.Helper.KeyGetter(action[:idx])
		if err == nil {
			rr.URL.Path = action[idx:] // 更新 URL.Path，去掉第一个目录
			rr.URL.RawPath = strings.TrimPrefix(rr.URL.RawPath, "/"+key)
			if router, exist := aa.Router[key]; exist {
				rr.Header.Set("X-Router-Key", action[:idx]) // 设置 X-Router-Key 头部信息
				router.ServeHTTP(rw, rr)
				return
			}
		} else if err != IngoreError {
			// 其他错误，返回 500 Internal Server Error
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	// 没有匹配到任何路由，返回 404 Not Found
	http.NotFound(rw, rr)
}

// SetRouterKeyByDir 从 URL.Path 中提取第一个目录作为 key，并设置到 X-Router-Key 头部信息中，同时更新 URL.Path 去掉第一个目录
func SetRouterKeyByDir(rr *http.Request) {
	if rr == nil {
		return
	}
	action := rr.URL.Path
	if len(action) > 0 && action[0] == '/' {
		action = action[1:]
	}
	if idx := strings.Index(action, "/"); idx > 0 {
		rkey := action[:idx]
		rr.Header.Set("X-Router-Key", rkey) // 设置 X-Router-Key 头部信息
		rr.URL.Path = action[idx:]          // 更新 URL.Path，去掉第一个目录
		rr.URL.RawPath = strings.TrimPrefix(rr.URL.RawPath, "/"+rkey)
	}
}

// -----------------------------------------------------------------------------------

// hst 识别 host，让后在进行 rdx 路由识别， 路由样式 http://{hst}.<domain>/...

// var _ Engine = (*HstRouter)(nil)

func NewHstRouter(svckit SvcKit) Engine {
	helper, _ := svckit.Get("rde-helper").(RdeHelper)
	return &HstRouter{
		name:   "zoo-hst",
		svckit: svckit,
		Helper: helper,
		Router: make(map[string]Engine),
	}
}

type HstRouter struct {
	name   string
	svckit SvcKit
	Helper RdeHelper
	Router map[string]Engine
}

func (aa *HstRouter) Name() string {
	return aa.name
}

func (aa *HstRouter) Handle(method, action string, handle HandleFunc) {
	if aa.Helper == nil {
		// 如果 Helper 为空，直接返回，不注册任何路由
		return
	}
	if G.Server.ApiRoot != "" {
		action = strings.TrimPrefix("/"+action, G.Server.ApiRoot+"/")
	}
	// 分拆 action 为 key 和 path
	key, path := "", ""
	if idx := strings.Index(action, "/"); idx > 0 {
		key, path = action[:idx], action[idx+1:]
	} else {
		key, path = action, ""
	}
	// 获取路由是否存在，不存在则创建一个新的路由并注册到 aa.Router 中，最后在路由上注册 handle
	router, exist := aa.Router[key]
	if !exist {
		router = aa.Helper.NewRouter(aa.svckit)
		aa.Router[key] = router
	}
	router.Handle(method, path, handle)
}

func (aa *HstRouter) ServeHTTP(rw http.ResponseWriter, rr *http.Request) {
	if aa.Helper == nil {
		// 如果 Helper 为空，直接返回 404 Not Found
		http.NotFound(rw, rr)
		return
	}
	if G.Server.ApiRoot != "" {
		rr.URL.Path = strings.TrimPrefix(rr.URL.Path, G.Server.ApiRoot+"/")
		rr.URL.RawPath = strings.TrimPrefix(rr.URL.RawPath, G.Server.ApiRoot+"/")
	}
	// 从 Host 中提取子域名作为 key
	host := rr.Host
	if idx := strings.Index(host, "."); idx > 0 {
		key, err := aa.Helper.KeyGetter(host[:idx])
		if err == nil {
			if router, exist := aa.Router[key]; exist {
				rr.Header.Set("X-Router-Key", host[:idx]) // 设置 X-Router-Key 头部信息
				router.ServeHTTP(rw, rr)
				return
			}
		} else if err != IngoreError {
			// 其他错误，返回 500 Internal Server Error
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	// 没有匹配到任何路由，返回 404 Not Found
	http.NotFound(rw, rr)
}

// SetRouterKeyByHst 从 Host 中提取子域名作为 key，并设置到 X-Router-Key 头部信息中，供后续路由识别使用
func SetRouterKeyByHst(rr *http.Request) {
	if rr == nil {
		return
	}
	host := rr.Host
	if idx := strings.Index(host, "."); idx > 0 {
		rkey := host[:idx]
		rr.Header.Set("X-Router-Key", rkey) // 设置 X-Router-Key 头部信息
	}
}

// -----------------------------------------------------------------------------------
