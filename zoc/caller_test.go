// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

package zoc_test

import (
	"fmt"
	"testing"

	"github.com/suisrc/zoo/zoc"
)

// go test -v z/zc/log_test.go -run Test_caller0

func Test_caller0(t *testing.T) {
	// 测试普通函数
	foo()

	// 测试值方法
	user := User{Name: "Alice"}
	user.GetName()

	// 测试指针方法
	userPtr := &User{Name: "Bob"}
	userPtr.SetName("Charlie")

	// 测试调用者信息
	fmt.Println("\n调用者信息:")
	info := zoc.GetCallerMethodInfo(0)
	fmt.Printf("%+v\n", zoc.ToStr(info))
}

// 普通函数
func foo() {
	info := zoc.GetCurrentMethodInfo()
	fmt.Printf("普通函数信息: %+v\n", zoc.ToStr(info))
}

// 测试结构体
type User struct {
	Name string
}

// 值方法
func (u User) GetName() string {
	info := zoc.GetCurrentMethodInfo()
	fmt.Printf("值方法信息: %+v\n", zoc.ToStr(info))
	return u.Name
}

// 指针方法
func (u *User) SetName(name string) {
	info := zoc.GetCurrentMethodInfo()
	fmt.Printf("指针方法信息: %+v\n", zoc.ToStr(info))
	u.Name = name
}
