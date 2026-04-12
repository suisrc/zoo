// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

package zoo

import (
	"flag"
	"fmt"
	"os"
	"time"
)

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
	LoadConfig(cfs)
	// running server
	zoo := &Zoo{}
	if zoo.ServeInit() {
		zoo.RunServe()
	}
}

// 打印版本信息
func PrintVersion() {
	Logn(AppName, Version, AppInfo, "PID:", os.Getpid())
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
