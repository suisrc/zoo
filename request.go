// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

package zoo

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

// 请求数据
func ReadBody[T any](rr *http.Request, rb T) (T, error) {
	return rb, json.NewDecoder(rr.Body).Decode(rb)
}

// 请求结构体， 特殊的请求体
type RaData struct {
	Atyp string `json:"type"`
	Data string `json:"data"`
}

// 请求数据
func ReadData(rr *http.Request) (*RaData, error) {
	return ReadBody(rr, &RaData{})
}

// 获取 traceID / 配置 traceID
func GetTraceID(request *http.Request) string {
	traceid := request.Header.Get("X-Request-Id")
	if traceid == "" {
		traceid = genStr("r", 32) // 创建请求ID, 用于追踪
		request.Header.Set("X-Request-Id", traceid)
	}
	return traceid
}

// 获取 reqType / 配置 reqType
func GetReqType(request *http.Request) string {
	reqtype := request.Header.Get("X-Request-Rt")
	if reqtype == "" {
		reqtype = G.Server.ReqXrtd
		if reqtype != "" {
			request.Header.Set("X-Request-Rt", reqtype)
		}
	}
	return reqtype
}

func GetRemoteIP(req *http.Request) string {
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		ip = strings.TrimSpace(strings.Split(ip, ",")[0])
		if ip == "" {
			ip = strings.TrimSpace(req.Header.Get("X-Real-Ip"))
		}
		if ip != "" {
			return ip
		}
	}
	if ip := req.Header.Get("X-Appengine-Remote-Addr"); ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// request token auth
func TokenAuth(token *string, handle HandleFunc) HandleFunc {
	// 需要验证令牌
	return func(ctx *Ctx) {
		if token == nil || *token == "" {
			handle(ctx) // auth pass
		} else if ktn := ctx.Request.Header.Get("Authorization"); ktn == "Token "+*token {
			handle(ctx) // auth succ
		} else {
			ctx.JSON(&Result{ErrCode: "invalid-token", Message: "无效的令牌"})
		}
	}
}

// merge multi func to one func
func MergeFunc(handles ...HandleFunc) HandleFunc {
	return func(ctx *Ctx) {
		for _, handle := range handles {
			handle(ctx)
			if ctx.IsAbort() {
				return
			}
		}
	}
}

// 响应数据
func WriteResp(rw http.ResponseWriter, ctype string, code int, data []byte) {
	h := rw.Header()
	// See https://go.dev/issue/66343.
	h.Del("Content-Length")
	// 设置 X-Content-Type-Options: nosniff 后，浏览器会严格遵循服务器返回的 Content-Type，不会尝试猜测资源类型。
	h.Set("X-Content-Type-Options", "nosniff")
	// 响应数据
	h.Set("Content-Type", ctype)
	rw.WriteHeader(code)
	rw.Write(data)
}
