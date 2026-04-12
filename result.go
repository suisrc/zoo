package zoo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var ResultEncoders = map[string]ResultEncoder{
	"2": EncodeJson2,
	"3": EncodeHtml3,
}

// ----------------------------------------------------------------------------

type Result2 struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	ErrCode string `json:"errorCode,omitempty"`
	Message string `json:"errorMessage,omitempty"`
	ErrShow int    `json:"showType,omitempty"`
	TraceID string `json:"traceId,omitempty"`
	Total   *int   `json:"total,omitempty"`
}

func (aa *Result2) Error() string {
	return fmt.Sprintf("[%v], %s, %s", aa.Success, aa.ErrCode, aa.Message)
}

func EncodeJson2(rr *http.Request, rw http.ResponseWriter, rs *Result) {
	// 转换结构
	aa := &Result2{}
	aa.Success = rs.Success
	aa.Data = rs.Data
	aa.TraceID = rs.TraceID
	if rs.ErrCode != "" {
		aa.ErrCode = rs.ErrCode
		aa.Message = rs.Message
	}
	if rs.ErrShow > 0 {
		aa.ErrShow = rs.ErrShow
	}
	if rs.Total != nil {
		aa.Total = rs.Total
	}
	// 响应结果
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	if rs.Status > 0 {
		rw.WriteHeader(rs.Status)
	}
	json.NewEncoder(rw).Encode(aa)
}

// ----------------------------------------------------------------------------

func EncodeHtml3(rr *http.Request, rw http.ResponseWriter, rs *Result) {
	if rs.Ctx == nil {
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.Write([]byte("template render error: request content not found"))
		return
	}
	tmpl := rs.TplKey
	if tmpl == "" {
		if rs.Success {
			tmpl = "success.html"
		} else {
			tmpl = "error.html"
		}
	}
	if rs.Status > 0 {
		rw.WriteHeader(rs.Status)
	}
	HTML0(rs.Ctx.SvcKit.Zoo(), rr, rw, rs, tmpl)
}
