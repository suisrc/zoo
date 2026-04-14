// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

// zoc: context

package zoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"text/template"
)

// 定义处理函数
type HandleFunc func(zrc *Ctx)

// 请求上下文内容
type Ctx struct {
	Ctx context.Context
	// Cancel func
	Cancel context.CancelFunc
	// All Module
	SvcKit SvcKit
	// Request action
	Action string
	// Request Cache
	Caches map[string]any
	// Params
	Params func(key string) string
	// Request
	Request *http.Request
	// Response
	Writer http.ResponseWriter
	// Trace ID
	TraceID string
	// X-Request-Rt
	ReqType string
	// flag router name
	_router string
	// flag action abort
	_abort bool
}

// 用于标记提前结束，不是强制的
func (ctx *Ctx) Abort() {
	ctx._abort = true
}

func (ctx *Ctx) IsAbort() bool {
	return ctx._abort
}

// 已 JSON 格式写出响应
func (ctx *Ctx) JSON(err error) {
	ctx._abort = true // rc.Abort()
	// 注意，推荐使用 JSON(rc, rs), 这里只是为了简化效用逻辑
	switch err := err.(type) {
	case *Result:
		JSON(ctx, err)
	default:
		JSON(ctx, &Result{ErrCode: "unknow-error", Message: err.Error()})
	}
}

// 已 HTML 模板格式写出响应
func (ctx *Ctx) HTML(tpl string, res any, hss int) {
	ctx._abort = true
	if ctx.TraceID != "" {
		ctx.Writer.Header().Set("X-Request-Id", ctx.TraceID)
	}
	if hss > 0 {
		ctx.Writer.WriteHeader(hss)
	}
	HTML0(ctx.SvcKit.Engine(), ctx.Request, ctx.Writer, res, tpl)
}

// 已 TEXT 模板格式写出响应
func (ctx *Ctx) TEXT(txt string, hss int) {
	ctx._abort = true
	if ctx.TraceID != "" {
		ctx.Writer.Header().Set("X-Request-Id", ctx.TraceID)
	}
	ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if hss > 0 {
		ctx.Writer.WriteHeader(hss) // 最后写状态码头
	}
	ctx.Writer.Write([]byte(txt))
}

// 已 BYTE 模板格式写出响应
func (ctx *Ctx) BYTE(bts io.Reader, hss int, cty string) {
	ctx._abort = true
	if ctx.TraceID != "" {
		ctx.Writer.Header().Set("X-Request-Id", ctx.TraceID)
	}
	if cty != "" {
		ctx.Writer.Header().Set("Content-Type", cty)
	}
	if hss > 0 {
		ctx.Writer.WriteHeader(hss) // 最后写状态码头
	}
	io.Copy(ctx.Writer, bts)
}

// 已 JSON 错误格式写出响应
func (ctx *Ctx) JERR(err error, hss int) {
	ctx._abort = true // rc.Abort()
	// 注意，推荐使用 JSON(rc, rs), 这里只是为了简化效用逻辑
	var res *Result
	switch err := err.(type) {
	case *Result:
		res = err
	default:
		res = &Result{ErrCode: "unknow-error", Message: err.Error()}
	}
	if hss > 0 {
		res.Status = hss
	}
	JSON(ctx, res)
}

// 创建上下文函数
func NewCtx(svckit SvcKit, request *http.Request, writer http.ResponseWriter, router string) *Ctx {
	action := GetAction(request.URL)
	ctx := &Ctx{SvcKit: svckit, Action: action, Caches: map[string]any{}, Request: request, Writer: writer}
	ctx._router = router
	ctx.Ctx, ctx.Cancel = context.WithCancel(context.Background())
	ctx.TraceID = GetTraceID(request)
	ctx.ReqType = GetReqType(request)
	return ctx
}

// 清理访问资源
func (ctx *Ctx) Clear() {
	ctx.Cancel()
	// 重点是清除指针，防止内存泄漏，
	// 因此如果是延迟或多线程处理时候，一定要 Clone Ctx, 否则无法在请求结束后使用
	ctx.Ctx = nil
	ctx.Cancel = nil
	ctx.SvcKit = nil
	ctx.Caches = nil
	ctx.Params = nil
	ctx.Request = nil
	ctx.Writer = nil
}

// 克隆上下文函数
func (ctx *Ctx) Clone(hasContext, hasRequest bool) *Ctx {
	clo := Ctx{}
	if hasContext {
		clo.Ctx = ctx.Ctx
		clo.Cancel = ctx.Cancel
	}
	clo.SvcKit = ctx.SvcKit
	clo.Action = ctx.Action
	clo.Caches = ctx.Caches
	clo.ReqType = ctx.ReqType
	clo.TraceID = ctx.TraceID
	clo._router = ctx._router
	// 拷贝参数
	if hasRequest {
		clo.Params = ctx.Params
		clo.Request = ctx.Request
		clo.Writer = ctx.Writer
		clo._abort = ctx._abort
	}
	return &clo
}

// 获取请求 action, 优先使用 query.action, 其次使用 path[1:] 作为 action
func GetAction(uu *url.URL) string {
	action := uu.Query().Get("action")
	if action == "" {
		rpath := uu.Path
		if len(rpath) > 0 {
			rpath = rpath[1:] // 删除前缀 '/'
		}
		action = rpath
	}
	return action
}

// 前缀匹配, 用于路径判断, 例如 /api/v1/user, pre=/api, 则返回 true
func HasPrepath(path string, epre string) bool {
	if elen := len(epre); elen == 0 {
		return true // 对比 epre 为空，直接返回 true
	} else if plen := len(path); plen == 0 {
		return false // 参考 path 为空，直接返回 false
	} else if plen < elen || path[:elen] != epre {
		return false // epre 不是 path 前缀，返回 false
	} else {
		// 1. 完全相等 2. epre 以 '/' 结尾， 3. path[elen] 是 '/'， 则返回 true
		return plen == elen || epre[elen-1] == '/' || path[elen] == '/'
	}
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

type ResultEncoder func(rr *http.Request, rw http.ResponseWriter, rs *Result)

// 定义响应结构体
type Result struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	ErrCode string `json:"errcode,omitempty"`
	Message string `json:"message,omitempty"`
	ErrShow int    `json:"errshow,omitempty"`
	TraceID string `json:"traceid,omitempty"`
	Total   *int   `json:"total,omitempty"`

	Ctx    *Ctx              `json:"-"`
	Status int               `json:"-"`
	Header map[string]string `json:"-"`
	TplKey string            `json:"-"`
}

func (aa *Result) Error() string {
	return fmt.Sprintf("[%v], %s, %s", aa.Success, aa.ErrCode, aa.Message)
}

// ----------------------------------------------------------------------------

// 响应 JSON 结果, 这是一个套娃，
func JSON(ctx *Ctx, res *Result) {
	res.Ctx = ctx
	// TraceID 可能不存在，如果不是 '' 则 PASS
	if res.TraceID == "" {
		res.TraceID = ctx.TraceID
	}
	if res.TraceID != "" {
		ctx.Writer.Header().Set("X-Request-Id", res.TraceID)
	}
	if !res.Success && res.ErrShow <= 0 {
		res.ErrShow = 1
	}
	// 响应头部
	if ctx.Request.Header != nil { // 设置响应头
		for k, v := range res.Header {
			ctx.Writer.Header().Set(k, v)
		}
	}
	// 响应结果
	if rfn, ok := ResultEncoders[ctx.ReqType]; ok {
		rfn(ctx.Request, ctx.Writer, res)
	} else {
		JSON0(ctx.Request, ctx.Writer, res)
	}
}

// 响应 JSON 结果: content-type http-status json-data
func JSON0(rr *http.Request, rw http.ResponseWriter, rs *Result) {
	// 响应结果
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	if rs.Status > 0 {
		rw.WriteHeader(rs.Status)
	}
	json.NewEncoder(rw).Encode(rs)
}

// 响应 HTML 模板结果: content-type http-status html-data
func HTML0(zg *Zoo, rr *http.Request, rw http.ResponseWriter, rs any, tp string) {
	// 响应结果
	if zg == nil {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.Write([]byte("template render error: server not found"))
	} else {
		err := zg.TplKit.Render(rw, tp, rs)
		if err != nil {
			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			rw.Write([]byte("template render error: " + err.Error()))
		} else {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
	}
}

// ----------------------------------------------------------------------------
// ----------------------------------------------------------------------------

// close function
type Closed func()

type TplCtx struct {
	Key string             // 模版编码
	Tpl *template.Template // 模版
	Err error              // 加载模版的异常
	Txt string             // 模版原始内容
	Lck sync.Mutex         // 模版锁
}

// 模版工具接口
type TplKit interface {
	Get(key string) *TplCtx
	Render(wr io.Writer, name string, data any) error
	Load(key string, str string) *TplCtx
	Preload(dir string) error
}

// 服务工具接口
type SvcKit interface {
	Engine() *Zoo                      // 获取模块管理器 *Zoo 接口
	Inject(obj any) SvcKit             // 注册服务 inject 使用 `svckit:"xxx"` 初始化服务
	Router(key string, hdl HandleFunc) // 注册路由
	Get(key string) any                // 获取服务
	Set(key string, val any) SvcKit    // 增加服务 val = nil 是卸载服务
	Map() map[string]any               // 服务列表, 注意，是副本
}

// 引擎接口, Engine, 不适用 Router 是为了和 其他 Router 名字上区分开。以便于支持多 Router 而不会出现冲突
type Engine interface {
	Name() string                                       // router engine name
	Handle(method, action string, handle HandleFunc)    // register router handle, [method]可能为"", [action]开头无"/"
	ServeHTTP(rw http.ResponseWriter, rr *http.Request) // http.HandlerFunc
}

// -----------------------------------------------------------------------------------

type Slice[T any] []T

// 追加一条数据到数据集中
func (s *Slice[T]) Add(vals ...T) {
	*s = append(*s, vals...)
}

// 删除满足条件的第一条数据
func (s *Slice[T]) Del(fn func(val T) bool) {
	if idx := slices.IndexFunc(*s, fn); idx >= 0 {
		*s = slices.Delete(*s, idx, idx+1)
	}
}

// 校验数据是否存在，存在替换，不存在追加
func (s *Slice[T]) Set(fn func(val T) bool, val T) {
	if idx := slices.IndexFunc(*s, fn); idx >= 0 {
		(*s)[idx] = val
	} else {
		*s = append(*s, val)
	}
}

// 获取满足条件的第一条数据
func (s *Slice[T]) Get(fn func(val T) bool) (val T, ok bool) {
	if idx := slices.IndexFunc(*s, fn); idx >= 0 {
		return (*s)[idx], true
	}
	var zero T
	return zero, false
}
