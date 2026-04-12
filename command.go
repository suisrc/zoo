// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

// 所有三方依赖，包括框架内的工具函数都在此文件中引入，方便管理和替换

package zoo

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/suisrc/zoo/zoc"
)

var (
	isDebug = zoc.IsDebug

	// 日志函数， 也可以直接使用 slog 包，这个包含文件和行号的追踪功能
	logf = zoc.Logf
	logn = zoc.Logn
	logz = zoc.Logz
	exit = zoc.Exit

	// 其他工具函数
	genStr   = zoc.GenStr
	funcInfo = zoc.FuncInfo
)

// 注册默认方法
func Initializ() {

	flag.Var(zoc.NewBoolVal(&(zoc.G.Debug)), "debug", "debug mode")
	flag.Var(zoc.NewBoolVal(&(zoc.G.Print)), "print", "print mode")
	flag.Var(zoc.NewBoolVal(&(G.Server.Fxser)), "fxser", "http header flag xser-*")
	flag.Var(zoc.NewBoolVal(&(G.Server.Local)), "local", "http server local mode")
	flag.StringVar(&(G.Server.Addr), "addr", "0.0.0.0", "http server addr")
	flag.IntVar(&(G.Server.Port), "port", 80, "http server Port")
	flag.IntVar(&(G.Server.Ptls), "ptls", 443, "https server Port")
	flag.BoolVar(&(G.Server.Dual), "dual", false, "running http and https server")
	flag.StringVar(&(G.Server.Engine), "eng", "map", "http server router engine")
	flag.StringVar(&(G.Server.ApiRoot), "api", "", "http server api root")
	flag.StringVar(&(G.Server.TplPath), "tpl", "", "templates folder path")
	flag.StringVar(&(G.Server.ReqXrtd), "xrt", "", "X-Request-Rt default value")

	// 注册配置函数
	Register("90-server", &G, RegisterHttpServe)
}

// 程序入口
func Execute(envpre, appname, version, appinfo string) {
	if envpre != "" {
		zoc.CFG_ENV = strings.ToUpper(envpre)
	}
	AppName, Version, AppInfo = appname, version, appinfo
	if len(os.Args) < 2 || strings.HasPrefix(os.Args[1], "-") {
		cmds["web"]() // run  def http server
		return        // wait for server stop
	}
	cmd := os.Args[1]
	if command, ok := cmds[cmd]; ok {
		// 修改命令参数
		os.Args = append(os.Args[:1], os.Args[2:]...)
		command() // run command
		// flag.Parse() > flag.CommandLine.Parse(os.Args[2:])
	} else {
		fmt.Println("unknown command:", cmd)
	}
}

// ----------------------------------------------------------------------------

var (
	// 应用配置列表，依据 key 排序，初始化顺序
	options = []Ref[string, func(SvcKit) Closed]{}
	optlock = sync.Mutex{} // 注册方法，全局锁即可
)

// 在 init 注册配置函数
func Register(key string, opts ...any) {
	for _, opt := range opts {
		if val, ok := opt.(func(SvcKit) Closed); ok {
			optlock.Lock() // 注册方法，全局锁即可
			defer optlock.Unlock()
			// options = append(options, Ref[string, OptionFunc]{Key: key, Val: opt})
			idx := slices.IndexFunc(options, func(opt Ref[string, func(SvcKit) Closed]) bool {
				return opt.Key > key
			})
			ref := Ref[string, func(SvcKit) Closed]{Key: key, Val: val}
			if idx < 0 {
				options = append(options, ref)
			} else {
				options = slices.Insert(options, idx, ref)
			}
		} else {
			zoc.Register(opt) // 其他类型，直接注册到 zoc 中
		}
	}
}

// ----------------------------------------------------------------------------

// Command Map Registry
var cmds = map[string]func(){
	"web":     RunHttpServe,
	"version": PrintVersion,
}

// 增加命令， 如果 cmd 为空，则删除该命令
func AddCmd(name string, cmd func()) {
	if cmd == nil {
		delete(cmds, name)
	} else {
		cmds[name] = cmd
	}
}

// 运行 http 服务
func RunHttpServe() {
	PrintVersion()
	Initializ()
	// parse command line arguments
	var cfs string
	flag.StringVar(&cfs, "c", "", "config file path")
	flag.Parse()
	// parse config file
	zoc.LoadConf(cfs)
	// running server
	zoo := &Zoo{}
	if zoo.ServeInit() {
		zoo.RunServe()
	}
}

// 打印版本信息
func PrintVersion() {
	logn(AppName, Version, AppInfo, "PID:", os.Getpid())
}

// ----------------------------------------------------------------------------
var HttpServeDef = true // 是否启动默认 http 服务? 这里不能通过配置，只能通过代码控制

func RegisterHttpServe(svc SvcKit) Closed {
	if !HttpServeDef {
		return nil // 不启动默认服务
	}
	if G.Server.Local {
		G.Server.Addr = "127.0.0.1"
	}
	engz := svc.Engine()
	if G.Server.Ptls > 0 && engz.TLSConf != nil {
		addr := fmt.Sprintf("%s:%d", G.Server.Addr, G.Server.Ptls)
		engz.Servers.Add(NewServer("(HTTPS)", engz, addr, engz.TLSConf))
	}
	if G.Server.Port > 0 && (engz.TLSConf == nil || G.Server.Dual) {
		addr := fmt.Sprintf("%s:%d", G.Server.Addr, G.Server.Port)
		engz.Servers.Add(NewServer("(HTTP1)", engz, addr, nil))
	}
	engz.AddRouter("healthz", Healthz) // 默认注册健康检查
	return nil
}

// 健康检查接口
func Healthz(ctx *Ctx) {
	ctx.JSON(&Result{Success: true, Data: time.Now().Format("2006-01-02 15:04:05")})
}

// favicon.ico
func Favicon(ctx *Ctx) {
	// 缓存1小时
	ctx.Writer.Header().Set("Cache-Control", "max-age=3600")
	ctx.Writer.Header().Set("Content-Type", "image/x-icon")
	ctx.Writer.Write([]byte{})
}

// ----------------------------------------------------------------------------

// 请求数据
func ReadForm[T any](rr *http.Request, rb T) (T, error) {
	return zoc.Map2ToStruct(rb, rr.URL.Query(), "form")
}
