package zoc_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/suisrc/zoo/zoc"
)

// go test -v z/zc/json_test.go -run Test_slice1

func Test_slice1(t *testing.T) {
	arrs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	arrs = zoc.SliceDelete(arrs, 9)
	println(zoc.ToStr(arrs))
	arrs = zoc.SliceDelete(arrs, 0)
	println(zoc.ToStr(arrs))
	arrs = zoc.SliceDelete(arrs, 1)
	println(zoc.ToStr(arrs))
	arrs = zoc.SliceDelete(arrs, 3)
	println(zoc.ToStr(arrs))
}

// go test -v z/zc/json_test.go -run Test_map1

func Test_map1(t *testing.T) {
	dmap := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "d",
				"d": 123,
				"j": "123",
				"k": int64(123),
				"l": float64(123),
				"m": float32(123),
				"n": int32(123),
				"e": "456",
				"f": 456.789,
				"g": uint16(0),
				"i": "N",
				"a": map[string]any{
					"j": "123",
				},
				"b": map[string]any{
					"j": "321",
				},
				"o": map[string]any{
					"j": 123,
				},
				"p": map[string]any{
					"j": true,
				},
				"q": map[string]any{
					"j": 123.456,
				},
			},
		},
	}

	t.Log("=================== ", zoc.MapAny(dmap, "a.b.e", false))
	t.Log("=================== ", zoc.MapAny(dmap, "a.b.f", false))
	t.Log("=================== ", zoc.MapAny(dmap, "a.b.g", true))
	t.Log("=================== ", zoc.MapAny(dmap, "a.b.h", false))
	t.Log("=================== ", zoc.MapAny(dmap, "a.b.i", true))
	zoc.MapNew(dmap, "a.b.x.y.-0.v", "123")
	t.Log("=================== ", zoc.MapAny(dmap, "a.b.x.y.0.v", 0))
	zoc.MapNew(dmap, "a.b.x.y.0.v", "456")
	t.Log("=================== ", zoc.MapAny(dmap, "a.b.x.y.0.v", 0))
	zoc.MapNew(dmap, "a.b.x.y.-0.-0.z", "123")
	t.Log("=================== ", zoc.MapAny(dmap, "a.b.x.y.0.0.z", 0))
	zoc.MapNew(dmap, "a.b.x.y.-0.-0.z", "123")
	zoc.MapNew(dmap, "a.b.x.y.1.-0.z", "789")
	zoc.MapNew(dmap, "a.b.x.y.1.-0.z", "567")
	zoc.MapNew(dmap, "a.b.x.y.-1.-0.z", "234")
	t.Log("=================== ", zoc.MapGet(dmap, "a.b.[.j=^*.2].j"))
	t.Log("=================== ", zoc.MapGet(dmap, "a.b.[.j=>122].j"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.[.=123]"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.[.='123']"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.[.=int.123]"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.[.=i64.123]"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.[.=f64.123]"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.[.=f32.123]"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.[.=i32.123]"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapSet(dmap, "a.b.d", nil))
	t.Log("=================== ", zoc.MapAny(dmap, "a.b.x.y.1.[.z=^*.6].z", 0))
}

// go test -v z/zc/json_test.go -run Test_map2

func Test_map2(t *testing.T) {
	dmap := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "d",
				"d": 123,
				"e": "456",
				"f": 456.789,
				"g": uint16(0),
				"i": "N",
				"a": map[string]any{
					"j": "123",
				},
				"b": map[string]any{
					"j": "321",
				},
				"x": map[string]any{
					"a": map[string]any{
						"j": "123",
					},
					"b": map[string]any{
						"j": "123",
					},
				},
				"z": map[string]any{
					"a": map[string]any{
						"j": "123",
					},
					"b": map[string]any{
						"j": "123",
					},
				},
			},
		},
	}
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.*"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.*"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*.j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.*.j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*.[.j=^*.2]"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.*.[.j=^*.2]"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*.[.j=^3]"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.*.[.j=^3]"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*.[.j=^1].j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.*.[.j=^1].j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.?.j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.?.j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a", "b", "?", "j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a", "b", "?", "j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a", "b", "*", "j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a", "b", "*", "j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*.*.j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.*.*.j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*.?.j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.*.?.j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.?.*.j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.?.*.j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.?.?.j"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.?.?.j"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.?.?.?"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.?.?.?"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.?.?.?.?"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.?.?.?.?"))
	t.Log("=================== ")
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.?.?.*"))
	t.Log("=================== ", zoc.MapGetsV2(dmap, "a.b.?.?.*"))
}

// go test -v z/zc/json_test.go -run Test_map3

func Test_map3(t *testing.T) {
	dmap := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "d",
				"d": 123,
				"e": "456",
				"f": 456.789,
				"g": uint16(0),
				"i": "N",
				"a": map[string]any{
					"j": "123",
				},
				"b": map[string]any{
					"j": "321",
				},
				"x": map[string]any{
					"a": map[string]any{
						"j": "123",
					},
					"b": map[string]any{
						"j": "123",
					},
				},
				"z": map[string]any{
					"a": map[string]any{
						"j": "123",
					},
					"b": map[string]any{
						"j": "123",
					},
				},
			},
		},
	}
	now := time.Now()
	for range 100_000 {
		zoc.MapGetsV2(dmap, "a.b.*.*.j")
		zoc.MapGetsV2(dmap, "a.b.*.?.j")
		zoc.MapGetsV2(dmap, "a.b.?.*.j")
	}
	t.Log("=================== ", time.Now().UnixMilli()-now.UnixMilli())
	now = time.Now()
	for range 100_000 {
		zoc.MapGets(dmap, "a.b.*.*.j")
		zoc.MapGets(dmap, "a.b.*.?.j")
		zoc.MapGets(dmap, "a.b.?.*.j")
	}
	t.Log("=================== ", time.Now().UnixMilli()-now.UnixMilli())
}

// go test -v z/zc/json_test.go -run Test_map4

func Test_map4(t *testing.T) {
	jsons := `
  {
    "apiVersion": "networking.k8s.io/v1",
    "kind": "Ingress",
    "metadata": {
      "name": "account-irs",
      "namespace": "rs-iam"
    },
    "spec": {
      "rules": [
        {
          "host": "",
          "http": {
            "paths": [
              {
                "backend": {
                  "service": {
                    "name": "end-fmes-adv-svc",
                    "port": {
                      "name": "http"
                    }
                  }
                },
                "path": "/api/adv",
                "pathType": "Prefix"
              },
              {
                "backend": {
                  "service": {
                    "name": "fnt-iam-account-m1-svc",
                    "port": {
                      "name": "http"
                    }
                  }
                },
                "path": "/m1/",
                "pathType": "Prefix"
              },
              {
                "backend": {
                  "service": {
                    "name": "fnt-account-svc",
                    "port": {
                      "name": "http"
                    }
                  }
                },
                "path": "/",
                "pathType": "Prefix"
              }
            ]
          }
        }
      ]
    }
  }`
	jsonv := map[string]any{}
	json.Unmarshal([]byte(jsons), &jsonv)
	// t.Log("=================== ", jsonv)
	t.Log("=================== ", zoc.MapGets(jsonv, "spec.rules.*.http.paths.*.backend.service[.name=^fnt-].name"))
	t.Log("=================== ", zoc.MapGets(jsonv, "spec.rules.*.http.paths.*.backend.service.name[.=^fnt-]"))
}

// go test -v z/zc/json_test.go -run Test_map5

func Test_map5(t *testing.T) {
	dmap := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{
					"q": map[string]any{
						"j": 123.456,
					},
				},
			},
		},
	}

	t.Log("=================== ", zoc.MapSet(dmap, "a.b.-0", map[string]any{}))
	t.Log("=================== ", zoc.MapSet(dmap, "a.b.0.m", 123))
	t.Log(zoc.ToStrJSON(dmap))
}

// go test -v z/zc/json_test.go -run Test_map6

func Test_map6(t *testing.T) {
	dmap := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{
					"q": map[string]any{
						"j": 123.456,
					},
				},
			},
		},
	}

	t.Log("=================== ", zoc.MapSet1(dmap, map[string]any{}, "a.b.-0"))
	t.Log("=================== ", zoc.MapSet1(dmap, 123456, "a.b.0.q.j"))
	t.Log("=================== ", zoc.MapSet1(dmap, 123456, "a.b.0.q.s"))
	t.Log(zoc.ToStrJSON(dmap))
	t.Log("=================== ", zoc.MapSets(dmap, 456789, "a.b.0.q.*"))
	t.Log(zoc.ToStrJSON(dmap))
	t.Log("=================== ", zoc.MapSets(dmap, 345678, "a.b.0.q.?"))
	t.Log(zoc.ToStrJSON(dmap))
	t.Log("=================== ", zoc.MapSets(dmap, 456789, "a.b.0.q.v.s.v"))
	t.Log("=================== ", zoc.MapSets(dmap, 456789, "a.b.0.v.-0.s"))
	t.Log("=================== ", zoc.MapSets(dmap, 123456, "a.b.0.v.-0.s"))
	t.Log(zoc.ToStrJSON(dmap))
	t.Log("=================== ", zoc.MapSets(dmap, 123456, "a.b.0.w.-0.s"))
	t.Log(zoc.ToStrJSON(dmap))
	t.Log("=================== ", zoc.MapSets(dmap, 123456, "a.b.0.w.-0.s"))
	t.Log(zoc.ToStrJSON(dmap))
}

// go test -v z/zc/json_test.go -run Test_map7

func Test_map7(t *testing.T) {
	dmap := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{
					"q": map[string]any{
						"j": 123.456,
					},
				},
			},
		},
	}

	t.Log("=================== ", zoc.MapSet1(dmap, map[string]any{"x": 456}, "a.b.-0"))
	t.Log("=================== ", zoc.MapGet1(dmap, "a.b.0.q.j"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.*"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.?"))
	t.Log("=================== ", zoc.MapGets(dmap, "a.b.1"))
	t.Log("=================== ", zoc.MapSet1(dmap, 123567, "a.b.*"))
	t.Log(zoc.ToStrJSON(dmap))
	t.Log("=================== ", zoc.MapSets(dmap, 123, "a.b.*"))
	t.Log(zoc.ToStrJSON(dmap))
}

// go test -v z/zc/json_test.go -run Test_map8

func Test_map8(t *testing.T) {
	dmap := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{
					"q": map[string]any{
						"j": 123.456,
					},
				},
			},
		},
	}
	t.Log("=================== ", zoc.MapSets(dmap, nil, "a.b.0.w.-0.s"))
	t.Log(zoc.ToStrJSON(dmap))
	t.Log("=================== ", zoc.MapNew(dmap, "a.b.0.w.-0.s", nil))
	t.Log(zoc.ToStrJSON(dmap))
}

// go test -v z/zc/json_test.go -run Test_map9

func Test_map9(t *testing.T) {
	dmap := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{
					"q": map[string]any{
						"j": 123.456,
					},
				},
				map[string]any{},
				map[string]any{},
			},
		},
	}
	t.Log("=================== ", zoc.MapDelEmp(dmap, "a.b.1"))
	t.Log(zoc.ToStr(dmap))
	t.Log("=================== ", zoc.MapDelSet[map[any]any](dmap, "a.b.1"))
	t.Log(zoc.ToStr(dmap))
	t.Log("=================== ", zoc.MapDelSet[map[string]any](dmap, "a.b.1"))
	t.Log(zoc.ToStr(dmap))
}
