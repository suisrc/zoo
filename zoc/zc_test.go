// Copyright 2026 suisrc. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/suisrc/zoo/blob/main/LICENSE.

package zoc_test

import (
	"log"
	"reflect"
	"testing"

	"github.com/suisrc/zoo/zoc"
)

var text = []byte(`
# 运行模式(debug:调试,test:测试,release:正式)
runmode = "debug"

# 启动时是否打印配置参数
print = true

[s.a.b.c]
a = "1"

[[s.a.b.d]]
a = "1"

[[s.a.b.d]]
a = "2"

[[s.a.b.e]]
c = "1"

[[s.a.b.e]]
c = "2"

[[s.b.b.e]]
n = "3"

[[s.b.b.e]]
n = "4"

[[s.b.a.e]]
n = "5"

[[s.c.a.e]]
n = "5"

[[s.c.a.e]]
n = "6"

[[s.c.a.e]]
n = "7"
`)

// ZOO_S_A_B_D_0_A=12 go test -v z/zc/zc_test.go -run Test_config
// ZOO_S_C_A_E_2_N=12 go test -v z/zc/zc_test.go -run Test_config
// ZOO_DEBUG=1 go test -v z/zc/zc_test.go -run Test_config

func load() {
	println("init load")
}

func Test_config(t *testing.T) {
	zoc.Register(&Data{})
	zoc.Register(func() { println("init other") })
	zoc.Register(load)
	log.Println("===========================loading")
	zoc.LoadConfig("zc_test.toml,zc_test1.toml")
	log.Println("===========================loading")
}

// MiddleConfig 中间件启动和关闭
type Data struct {
	Debug   bool `default:"false"`
	Logger  bool `default:"true"`
	RunMode string
	B       struct {
		A *struct {
			B struct {
				D []struct {
					C int `default:"5" json:"a"`
				}
				E []*struct {
					C int `default:"6"`
				}
			}
		}
		B map[string]*struct {
			E []struct {
				C int `default:"5" json:"n"`
			}
		}
		C map[string]struct {
			E []struct {
				C int `default:"5" json:"n"`
			}
		}
	} `json:"s"`
}

// go test -v z/zc/zc_test.go -run Test_toml1

func Test_toml1(t *testing.T) {
	data := zoc.NewTOML(text).Map()
	log.Println(zoc.ToStrJSON(data))
}

// go test -v z/zc/zc_test.go -run Test_toml2

func Test_toml2(t *testing.T) {
	data := &Data{}
	zoc.NewTOML(text).Decode(data, "json")
	log.Println(zoc.ToStrJSON(data))
	// log.Println("===========================")
	// smap := zoc.ToMap(data, "json", true)
	// log.Println(zoc.ToStrJSON(smap))
}

// go test -v z/zc/zc_test.go -run Test_tags

func Test_tags(t *testing.T) {
	data := &Data{}
	zoc.NewTOML(text).Decode(data, "json")
	// tags, _, _ := zoc.ToTagMap(data, "json", true, nil)
	tags := zoc.ToTagVal(data, "json")
	log.Println("===========================")
	log.Println(zoc.ToStrJSON(tags))
	log.Println("===========================")
	tags[len(tags)-1].Value.Set(reflect.ValueOf(12))
	tags[0].Value.Set(reflect.ValueOf(true))
	log.Println(zoc.ToStrJSON(data))
}

// go test -v z/zc/zc_test.go -run Test_StrToArr

func Test_StrToArr(t *testing.T) {
	str := `["a","b"]`
	arr := zoc.ToStrArr(str)
	log.Println(arr)

	str = `[1, 2]`
	arr = zoc.ToStrArr(str)
	log.Println(arr)

	str = `[a, b]`
	arr = zoc.ToStrArr(str)
	log.Println(arr)

	str = `['a', 'b']`
	arr = zoc.ToStrArr(str)
	log.Println(arr)

	str = `123.123`
	aaa := zoc.ToStrOrArr(str)
	log.Println(aaa)

	str = `"123.123"`
	aaa = zoc.ToStrOrArr(str)
	log.Println(aaa)
	log.Println("======================")
	str = `["x", 1, true, 'y', "a'b", 'c"d', "e\\f"]`
	arr = zoc.ToStrArr(str)
	for _, v := range arr {
		log.Println(v)
	}
	log.Println("======================")
	str = `["x", 1, true, 'y', "a'b", 'c"d', "e\\f"]`
	aaa = zoc.ToStrOrArr(str)
	log.Println(aaa)

	log.Println("======================")
	str = `["x", "a'b", "e\\f"]`
	aaa = zoc.ToStrOrArr(str)
	log.Println(aaa)

	log.Println("======================")

}

func Test_arr(t *testing.T) {
	arr := []string{"a", "b", "c"}
	log.Println(arr)
	// log.Println(arr[:-1])
}

// go test -v z/zc/zc_test.go -run Test_ToBasicVal

func Test_ToBasicVal(t *testing.T) {
	str := []string{"123", "456"}
	var val any
	var err error
	val, err = zoc.ToBasicValue(reflect.TypeOf(float32(0)), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf(float64(0)), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf(int(0)), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf(int8(0)), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf(int32(0)), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf(int64(0)), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf(""), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf([]byte("")), str)
	log.Printf("=======%#v | %v\n", string(val.([]byte)), err)
	val, err = zoc.ToBasicValue(reflect.TypeOf(str), str)
	log.Printf("=======%#v | %v\n", val, err)

	val, err = zoc.ToBasicValue(reflect.TypeOf([]int{}), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf([]int32{}), str)
	log.Printf("=======%#v | %v\n", val, err)
	val, err = zoc.ToBasicValue(reflect.TypeOf([][]byte{}), str)
	log.Printf("=======%#v | %v\n", string(val.([][]byte)[0]), err)

}

type RecordDat0 struct {
	Data0 string
}

type RecordDat1 struct {
	Data1 string
	RecordDat0
}

type Record struct {
	RecordDat1
	NameKey string
	AgeKey  int

	DataKey RecordData
	DataKe2 *RecordData
}

func (r Record) MarshalJSON() ([]byte, error) {
	return zoc.ToJsonBytes(&r, "json", zoc.LowerFirst, false)
}

type RecordData struct {
	NameKey string
	AgeKey  int
}

func (r RecordData) MarshalJSON() ([]byte, error) {
	return zoc.ToJsonBytes(&r, "json", zoc.Camel2Case, false)
}

// go test -v z/zc/zc_test.go -run Test_ToJson

func Test_ToJson(t *testing.T) {
	record := Record{
		NameKey: "x",
		AgeKey:  12,
		DataKey: RecordData{
			NameKey: "",
			AgeKey:  13,
		},
	}
	log.Println(zoc.ToStrJSON(record))
}
