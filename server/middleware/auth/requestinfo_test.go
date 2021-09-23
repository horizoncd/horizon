package auth

import (
	"g.hz.netease.com/horizon/util/sets"
	"testing"
)


func TestRequestInfo(t *testing.T) {
	requestInfoFactory := RequestInfoFactory{
		APIPrefixes: sets.NewString("apis", "api"),
	}
}