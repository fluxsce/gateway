package middleware

import (
	"net/http"
	"testing"

	"gateway/web/utils/constants"
)

func TestMustChangePwdAllowed(t *testing.T) {
	root := constants.APIRoot + "/user"
	if !mustChangePwdAllowed(http.MethodPut, root+"/password") {
		t.Fatal("改密接口应放行")
	}
	if !mustChangePwdAllowed(http.MethodPost, root+"/logout") {
		t.Fatal("登出应放行")
	}
	if !mustChangePwdAllowed(http.MethodGet, root+"/userinfo") {
		t.Fatal("userinfo 应放行")
	}
	if mustChangePwdAllowed(http.MethodPost, constants.APIRoot+"/hub0002/queryUsers") {
		t.Fatal("业务接口不应放行")
	}
}
