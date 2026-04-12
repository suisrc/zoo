// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

package zoc_test

import (
	"flag"
	"testing"

	"github.com/suisrc/zoo/zoc"
)

type VV struct {
	V1 string
	V2 int
	V3 bool
	V4 []string
	V5 map[string]string
}

// go test -v z/zc/flag_test.go -run Test_flag

func Test_flag(t *testing.T) {

	config := VV{
		V1: "v1",
	}
	ff := &flag.FlagSet{}

	ff.StringVar(&config.V1, "v1", "", "config v1")
	ff.Var(zoc.NewStrArr(&config.V4, []string{"1", "2"}), "v4", "config v4")
	ff.Var(zoc.NewStrMap(&config.V5, map[string]string{"k1": "v2", "k2": "v3"}), "v5", "config v5")

	ff.Parse([]string{"-v1", "123", "-v4", "3,4,5,6", "-v5", "k3=v4,k4=v5,k5,k6="})
	t.Log("==", zoc.ToStr(config))
}
