// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

package zoo

import (
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

// service 管理工具

var _ SvcKit = (*svcz)(nil)

type svcz struct {
	engine *Zoo
	svcmap map[string]any
	typmap map[reflect.Type]any
	svclck sync.RWMutex
}

func NewSvcKit(engine *Zoo) SvcKit {
	svckit := &svcz{
		engine: engine,
		svcmap: make(map[string]any),
		typmap: make(map[reflect.Type]any),
	}
	svckit.svcmap["svckit"] = svckit
	svckit.typmap[reflect.TypeFor[*svcz]()] = svckit
	return svckit
}

func (aa *svcz) Engine() *Zoo {
	return aa.engine
}

func (aa *svcz) Router(key string, hdl HandleFunc) {
	aa.engine.AddRouter(key, hdl)
}

func (aa *svcz) Get(key string) any {
	aa.svclck.RLock()
	defer aa.svclck.RUnlock()
	return aa.svcmap[key]
}

func (aa *svcz) Set(key string, val any) SvcKit {
	aa.svclck.Lock()
	defer aa.svclck.Unlock()
	if val != nil {
		// create or update
		aa.svcmap[key] = val
		aa.typmap[reflect.TypeOf(val)] = val
	} else {
		// delete
		val := aa.svcmap[key]
		if val != nil {
			delete(aa.svcmap, key)
			// delete value by type
			for kk, vv := range aa.typmap {
				if vv == val {
					delete(aa.typmap, kk)
					break
				}
			}
		}
	}
	return aa
}

func (aa *svcz) Map() map[string]any {
	aa.svclck.RLock()
	defer aa.svclck.RUnlock()
	ckv := make(map[string]any)
	maps.Copy(ckv, aa.svcmap)
	return ckv
}

func (aa *svcz) toInjName(tType, tField string) string {
	name := fmt.Sprintf("%s.%s", tType, tField)
	if size := len(name); size < 36 {
		name += strings.Repeat(" ", 36-size)
	}
	return name
}

func (aa *svcz) Inject(obj any) SvcKit {
	aa.svclck.RLock()
	defer aa.svclck.RUnlock()
	// 构建注入映射
	tType := reflect.TypeOf(obj).Elem()
	tElem := reflect.ValueOf(obj).Elem()
	for i := 0; i < tType.NumField(); i++ {
		tField := tType.Field(i)
		tagVal := tField.Tag.Get("svckit")
		if tagVal == "" || tagVal == "-" {
			continue // 忽略
		}
		if tagVal == "type" || tagVal == "auto" {
			// 通过 `svckit:'type/auto'` 中的接口匹配注入
			found := false
			for vType, value := range aa.typmap {
				if tField.Type == vType || // 属性是一个接口，判断接口是否可以注入
					tField.Type.Kind() == reflect.Interface && vType.Implements(tField.Type) {
					tElem.Field(i).Set(reflect.ValueOf(value))
					if isDebug() {
						logf("[_svckit_]: [inject] %s <- %s\n", aa.toInjName(tType.String(), tField.Name), vType)
					}
					found = true
					break
				}
			}
			if !found {
				errstr := fmt.Sprintf("[_svckit_]: [inject] %s <- %s.(type) error, service not found", //
					aa.toInjName(tType.String(), tField.Name), tField.Type)
				if isDebug() {
					logn(errstr)
				} else {
					exit(errstr) // 生产环境，注入失败，则 panic
				}
			}
		} else {
			// 通过 `svckit:'(name)'` 中的 (name) 注入
			val := aa.svcmap[tagVal]
			if val == nil {
				errstr := fmt.Sprintf("[_svckit_]: [inject] %s <- %s.(name) error, service not found", //
					aa.toInjName(tType.String(), tField.Name), tagVal)
				if isDebug() {
					logn(errstr)
				} else {
					exit(errstr) // 生产环境，注入失败，则 panic
				}
				continue
			}
			tElem.Field(i).Set(reflect.ValueOf(val))
			if isDebug() {
				logf("[_svckit_]: [inject] %s <- %s\n", aa.toInjName(tType.String(), tField.Name), reflect.TypeOf(val))
			}
		}
	}
	return aa
}

// -----------------------------------------------------------------------------------

// 可获取有权限字段
func FieldValue(target any, field string) any {
	val := reflect.ValueOf(target)
	return val.Elem().FieldByName(field).Interface()
}

// 可设置字段值
func FieldSetVal(target any, field string, value any) {
	val := reflect.ValueOf(target)
	val.Elem().FieldByName(field).Set(reflect.ValueOf(value))
}

// 获取字段, 可夸包获取私有字段
// 闭包原则，原则上不建议使用该方法，因为改方法是在破坏闭包原则
func FieldValue_(target any, field string) any {
	val := reflect.ValueOf(target)
	vap := unsafe.Pointer(val.Elem().FieldByName(field).UnsafeAddr())
	return *(*any)(vap)
}

// -----------------------------------------------------------------------------------

// 获取 target 中每个字段的属性，注入和 value 属性的字段
// 这只是一个演示的例子，实际开发中，请使用 SvcKit 模块
func FieldInject(target any, value any, tag string, debug bool) bool {
	vType := reflect.TypeOf(value)
	tType := reflect.TypeOf(target).Elem()
	tElem := reflect.ValueOf(target).Elem()
	for i := 0; i < tType.NumField(); i++ {
		tField := tType.Field(i)
		tagVal := tField.Tag.Get(tag)
		if tagVal != "type" && tagVal != "auto" {
			continue // `"tag":"type/auto"` 才可以通过类型注入
		}
		// 判断 vType 是否实现 tField.Type 的接口 // 属性是一个接口，判断接口是否可以注入
		if tField.Type == vType || //
			tField.Type.Kind() == reflect.Interface && vType.Implements(tField.Type) {
			tElem.Field(i).Set(reflect.ValueOf(value))
			if debug {
				logf("[_inject_]: [succ] %s.%s <- %s", tType, tField.Name, vType)
			}
			return true // 注入成功
		}
	}
	if debug {
		logf("[_inject_]: [fail] %s not found field.(%s)", tType, vType)
	}
	return false
}

// -----------------------------------------------------------------------------------

// GET http method
func GET(key string, hdl HandleFunc, svc SvcKit) {
	svc.Router(http.MethodGet+" "+key, hdl)
}

// POST http method
func POST(key string, hdl HandleFunc, svc SvcKit) {
	svc.Router(http.MethodPost+" "+key, hdl)
}

/**
 * 注册服务, key 必须唯一, 如果 key 为空， 使用 val.(type).Name() 作为 key
 * @param kit 服务容器
 * @param inj 自动注入
 * @param key 服务 key
 * @param val 服务实例
 */
func RegKey[T any](kit SvcKit, inj bool, key string, val T) T {
	if key == "" {
		key = reflect.TypeOf(val).Elem().Name()
	}
	kit.Set(key, val)
	if inj {
		kit.Inject(val) // 自动注入， 可以注入自己
	}
	return val
}

// 注册服务， name = val.(type)
func RegSvc[T any](kit SvcKit, val T) T {
	key := reflect.TypeOf(val).Elem().Name()
	kit.Set(key, val)
	return val
}

// 注册服务， name = val.(type)， 并自动初始化 val 实体
func Inject[T any](kit SvcKit, val T) T {
	key := reflect.TypeOf(val).Elem().Name()
	kit.Set(key, val).Inject(val) // 自动注入， 可以注入自己
	return val
}
