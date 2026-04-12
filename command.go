// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

package zoo

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/suisrc/zoo/zoc"
)

// 程序入口
func Execute(appname, version, appinfo string) {
	AppName, Version, AppInfo = appname, version, appinfo
	if len(os.Args) < 2 || strings.HasPrefix(os.Args[1], "-") {
		CMD["web"]() // run  def http server
		return       // wait for server stop
	}
	cmd := os.Args[1]
	if command, ok := CMD[cmd]; ok {
		// 修改命令参数
		os.Args = append(os.Args[:1], os.Args[2:]...)
		command() // run command
		// flag.Parse() > flag.CommandLine.Parse(os.Args[2:])
	} else {
		fmt.Println("unknown command:", cmd)
	}
}

// Command Map Registry
var CMD = map[string]func(){
	"web":     RunHttpServe,
	"version": PrintVersion,
}

func RunHttpServe() {
	PrintVersion()
	Initializ()
	// parse command line arguments
	var cfs string
	flag.StringVar(&cfs, "c", "", "config file path")
	flag.Parse()
	// parse config file
	zoc.LoadConfig(cfs)
	// running server
	zoo := &Zoo{}
	if zoo.ServeInit() {
		zoo.RunServe()
	}
}

func PrintVersion() {
	Logn(AppName, Version, AppInfo, "PID:", os.Getpid())
}
