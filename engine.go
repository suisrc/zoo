// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

// zoo 核心内容，为简约而生

package zoo

import (
	"cmp"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	AppName = "zoo"
	Version = "v0.0.0"
	AppInfo = "(https://github.com/suisrc/zoo.git)"

	G = new(struct {
		Server EngineConfig
	})

	IngoreError = errors.New("ignore error")

	// 路由构建器
	Engines = map[string]func(SvcKit) Engine{
		"map": NewMapRouter,
		"mux": NewMuxRouter,
		"dir": NewDirRouter,
		"hst": NewHstRouter,
	}
)

// 默认配置， Engine 配置需要内嵌该结构体
type EngineConfig struct {
	Fxser   bool   `json:"xser"` // 标记 xser 头部信息
	Local   bool   `json:"local"`
	Addr    string `json:"addr" default:"0.0.0.0"`
	Port    int    `json:"port" default:"80"`
	Ptls    int    `json:"ptls" default:"443"`
	Dual    bool   `json:"dual"`   // http and https
	Engine  string `json:"engine"` // router engine
	ApiRoot string `json:"root"`   // root api root
	TplPath string `json:"tpl"`    // templates folder path
	ReqXrtd string `json:"xrt"`    // X-Request-Rt default value, 1: zoo, 2: ali, 3: html
}

// -----------------------------------------------------------------------------------

var _ http.Handler = (*Zoo)(nil)

// 默认服务实体
type Zoo struct {
	Servers Slice[Server] // 接口服务模块列表
	Closeds Slice[Closed] // 模块关闭函数列表
	TLSConf *tls.Config   // Certificates, GetCertificate

	Engine Engine // 路由引擎
	SvcKit SvcKit // 服务工具
	TplKit TplKit // 模版工具
	_abort bool   // 终止标记
}

// -----------------------------------------------------------------------------------

// 默认相应函数 http.HandlerFunc(zoo.ServeHTTP)
func (aa *Zoo) ServeHTTP(rw http.ResponseWriter, rr *http.Request) {
	if IsDebug() {
		Logf("[_request]: [%s] %s %s\n", aa.Engine.Name(), rr.Method, rr.URL.String())
	}
	if G.Server.Fxser {
		rw.Header().Set("Xser-Routerz", aa.Engine.Name())
		rw.Header().Set("Xser-Version", AppName+":"+Version)
	}
	aa.Engine.ServeHTTP(rw, rr)
}

// 服务初始化
func (aa *Zoo) ServeInit() bool {
	aa.Servers = Slice[Server]{}
	aa.Closeds = Slice[Closed]{}
	if aa.SvcKit == nil {
		aa.SvcKit = NewSvcKit(aa)
	}
	if builder, ok := Engines[G.Server.Engine]; !ok {
		Logf("[_router_]: router not found by [-eng %s]\n", G.Server.Engine)
		return false
	} else {
		aa.Engine = builder(aa.SvcKit)
		Logf("[_router_]: build %s.router by [-eng %s]\n", aa.Engine.Name(), G.Server.Engine)
	}
	if aa.TplKit == nil {
		aa.TplKit = NewTplKit(aa.SvcKit)
		if G.Server.TplPath != "" {
			err := aa.TplKit.Preload(G.Server.TplPath)
			if err != nil {
				Logf("[_tplkit_]: Preload error: %v\n", err)
			}
		}
	}
	// -----------------------------------------------
	Logn("[register]: register server options...")
	for _, opt := range options {
		if opt.Val == nil {
			continue
		}
		if IsDebug() {
			ekey := opt.Key
			if size := len(ekey); size < 42 {
				ekey += " " + strings.Repeat("-", 41-size)
			}
			Logf("[register]: %s", ekey)
		}
		cls := opt.Val(aa.SvcKit)
		if cls != nil {
			aa.Closeds.Add(cls)
		}
		if aa._abort {
			Logn("[register]: serve already stop! exit...")
			return false // 退出
		}
	}
	slices.Reverse(aa.Closeds) // 倒序, 后进先出
	return true
}

// 增加处理函数
// @param key: [method:]action, 如果 method 为空，则默认为 所有请求
func (aa *Zoo) AddRouter(key string, handle HandleFunc) {
	if key == "" {
		if IsDebug() {
			Logf("[_handle_]: %32s  %p  %s\n", "/", handle, GetFuncInfo(handle))
		}
		aa.Engine.Handle("", "", handle)
		return
	}
	// 解析 method 和 action
	method, action, found := key, "", false
	if i := strings.IndexAny(key, " \t"); i >= 0 {
		method, action, found = key[:i], strings.TrimLeft(key[i+1:], " \t"), true
	}
	if !found {
		action = method
		method = ""
	}
	if len(action) > 0 && action[0] == '/' { // 去除 action 前 /
		action = action[1:]
	}
	if G.Server.ApiRoot != "" { // 补充 api root
		// action = filepath.Join(G.Server.ApiRoot, action)
		action = G.Server.ApiRoot + "/" + action
		if action[0] == '/' {
			action = action[1:]
		}
	}
	if method != "" {
		method = strings.ToUpper(method)
	}

	if IsDebug() { // log for debug
		Logf("[_handle_]: %32s  %p  %s\n", method+" /"+action, handle, GetFuncInfo(handle))
	}
	aa.Engine.Handle(method, action, handle)
}

// 服务终止，注意，这里只会终止模版，不会终止服务， 终止服务，需要调用 hsv.Shutdown
func (aa *Zoo) ServeStop(err ...string) {
	if len(err) > 0 {
		Logz(1, "[_server_]: serve stop,", strings.Join(err, " "))
	}
	if aa._abort {
		return
	}
	aa._abort = true
	if aa.Closeds != nil {
		for _, cls := range aa.Closeds {
			cls() // 模块关闭
		}
	}
	Logn("[_server_]: services have been terminated")
}

// 启动 HTTP 服务
func (aa *Zoo) RunServe() {
	// 停止业务模块， 先停服务，后停模块
	defer aa.ServeStop()
	// 启动HTTP服务， 并可优雅的终止
	for _, srv := range aa.Servers {
		if srv != nil {
			Logn("[_server_]: http server booting... linsten:", srv.Name(), srv.Addr())
			go srv.RunServe()
		}
	}
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	aa.WaitFor()
}

// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
func (aa *Zoo) WaitFor() {
	if len(aa.Servers) == 0 {
		Logn("[_server_]: no server to wait for, exit...")
		return
	}
	ssc := make(chan os.Signal, 1)
	signal.Notify(ssc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-ssc
	Logn("[_server_]: services is shutting down...")
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, srv := range aa.Servers {
		if srv != nil {
			Logn("[_server_]: http server stoping...", srv.Name())
			if err := srv.Shutdown(ctx); err != nil {
				Logn("[_server_]: http server shutdown error:", srv.Name(), err)
			}
		}
	}
}

// -----------------------------------------------------------------------------------

type Server interface {
	Name() string
	Addr() string
	RunServe()
	Shutdown(ctx context.Context) error
}

func NewServer(name string, handler http.Handler, addr string, conf *tls.Config) Server {
	return &servez{Server: http.Server{Handler: handler, Addr: addr, TLSConfig: conf}, ErrExit: true, SrvName: name}
}

type servez struct {
	http.Server
	ErrExit bool
	SrvName string
}

func (srv *servez) Name() string {
	return srv.SrvName
}

func (srv *servez) Addr() string {
	return srv.Server.Addr
}

func (srv *servez) RunServe() {
	if srv.Server.TLSConfig != nil {
		if err := srv.Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			if srv.ErrExit {
				Exit(fmt.Sprintf("[_server_]: server exit error: %s\n", err))
			} else {
				Logn(fmt.Sprintf("[_server_]: server error: %s\n", err))
			}
		}
	} else if err := srv.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		if srv.ErrExit {
			Exit(fmt.Sprintf("[_server_]: server exit error: %s\n", err))
		} else {
			Logn(fmt.Sprintf("[_server_]: server error: %s\n", err))
		}
	}
}

// -----------------------------------------------------------------------------------

// 创建指针
func Ptr[T any](v T) *T {
	return &v
}

// 键值对
type Ref[K cmp.Ordered, T any] struct {
	Key K
	Val T
}

type BufferPool interface {
	Get() []byte
	Put([]byte)
}

// NewBufferPool 初始化缓冲池
// defCap: 新缓冲区的默认容量（如32KB）
// maxCap: 允许归还的最大容量（如1MB）
func NewBufferPool(defCap, maxCap int) BufferPool {
	if defCap <= 0 {
		defCap = 32 * 1024 // 默认32KB, 现代CPU L1缓存通常为32KB/核
	}
	if maxCap <= 0 {
		maxCap = 1024 * 1024 // 默认1MB
	}
	return &BufferPoolDef{
		defCap: defCap,
		maxCap: maxCap,
		pool: &sync.Pool{
			New: func() any {
				// 创建默认容量的空字节切片（len=0，cap=defaultCap）
				return make([]byte, 0, defCap)
			},
		},
	}
}

// BufferPoolDef 字节缓冲池：基于sync.Pool实现
type BufferPoolDef struct {
	pool   *sync.Pool
	maxCap int // 允许归还的最大缓冲区容量（避免超大缓冲区占用内存）
	defCap int // 新创建缓冲区的默认容量
}

// Get 获取缓冲区：从池取出或创建新缓冲区
func (p *BufferPoolDef) Get() []byte {
	return p.pool.Get().([]byte)
}

// Put 归还缓冲区：重置后放回池（若容量超过maxCap则丢弃）
func (p *BufferPoolDef) Put(buf []byte) {
	// 1. 检查缓冲区容量是否超过限制
	if cap(buf) > p.maxCap {
		buf = nil
		return
	}
	// 2. 重置缓冲区：保留容量，清空内容（len=0）
	p.pool.Put(buf[:0])
}
