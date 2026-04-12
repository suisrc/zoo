package zoo

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// template: 模版管理工具

var (
	ErrTplNotFound = errors.New("tpl not found")
)

var _ TplKit = (*TplKitDef)(nil)

type TplKitDef struct {
	tpls map[string]*TplCtx // 所有模版集合
	lock sync.RWMutex       // 读写锁

	FuncMap template.FuncMap // 支持链式调用
}

func NewTplKit(server *Zoo) TplKit {
	return &TplKitDef{
		tpls: make(map[string]*TplCtx),
	}
}

func (aa *TplKitDef) Get(key string) *TplCtx {
	aa.lock.RLock()
	defer aa.lock.RUnlock()
	return aa.tpls[key]
}

func (tk *TplKitDef) Render(wr io.Writer, name string, data any) error {
	tpl := tk.Get(name)
	if tpl == nil {
		return ErrTplNotFound
	} else if tpl.Err != nil {
		return tpl.Err
	}
	return tpl.Tpl.Execute(wr, data)
}

func (aa *TplKitDef) Load(key string, str string) *TplCtx {
	aa.lock.Lock()
	defer aa.lock.Unlock()
	if tpl, ok := aa.tpls[key]; ok {
		return tpl
	}
	tpl := &TplCtx{}
	tpl.Key = key
	tpl.Txt = str
	tpl.Tpl, tpl.Err = template.New(tpl.Key).Parse(tpl.Txt)
	if tpl.Err == nil {
		tpl.Tpl.Funcs(aa.FuncMap)
	}
	aa.tpls[tpl.Key] = tpl
	return tpl
}

func (aa *TplKitDef) Preload(dir string) error {
	aa.lock.Lock()
	defer aa.lock.Unlock()
	// 读取 dir 文件夹中 所有的 *.html 文件
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".html") {
			return nil
		}
		key := path
		if idx := strings.IndexRune(path, '/'); idx >= 0 {
			key = path[idx+1:]
		}
		// 读取文件内容
		txt, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tpl := &TplCtx{}
		tpl.Key = key
		tpl.Txt = string(txt)
		tpl.Tpl, tpl.Err = template.New(tpl.Key).Parse(tpl.Txt)
		if tpl.Err == nil {
			tpl.Tpl.Funcs(aa.FuncMap)
		}
		aa.tpls[tpl.Key] = tpl
		if IsDebug() {
			Logf("[_preload]: [tplkit] %s", tpl.Key)
		}
		return nil
	})
}
