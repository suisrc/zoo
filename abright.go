package zoo

import (
	"encoding/hex"
	"flag"
	"net/http"

	"github.com/suisrc/zoo/zoc"
)

// 所有三方依赖，包括框架内的工具函数都在此文件中引入，方便管理和替换

var (
	IsDebug = zoc.IsDebug
	// 日志函数， 也可以直接使用 slog 包，这个包含文件和行号的追踪功能
	Logf = zoc.Logf
	Logn = zoc.Logn
	Logz = zoc.Logz
	Exit = zoc.Exit
	// Deprecated: 已于v0.5.1中废弃 保留只是为了兼容旧版本，实际调用 Logn
	Println = zoc.Logn
	// Deprecated: 已于v0.5.1中废弃 保留只是为了兼容旧版本，实际调用 Logf
	Printf = zoc.Logf

	// 其他工具函数
	Config    = zoc.Register
	ToStr     = zoc.ToStr
	HexStr    = hex.EncodeToString
	GenStr    = zoc.GenStr
	GenUUIDv4 = zoc.GenUUIDv4
	UnicodeTo = zoc.UnicodeToRunes

	GetHostname  = zoc.GetHostname
	GetNamespace = zoc.GetNamespace
	GetLocAreaIp = zoc.GetLocAreaIp
	GetServeName = zoc.GetServeName
	GetFuncInfo  = zoc.GetFuncInfo
)

// 注册默认方法
func Initializ() {
	// 注册配置函数
	Config(G)

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

	//  register default serve
	Register("90-server", RegisterHttpServe)
}

// 请求数据
func ReadForm[T any](rr *http.Request, rb T) (T, error) {
	return zoc.Map2ToStruct(rb, rr.URL.Query(), "form")
}
