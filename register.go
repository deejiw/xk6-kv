package kv

import (
	"github.com/deejiw/xk6-kv/kv"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/kv", kv.New())
}
