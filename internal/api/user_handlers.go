package api

import (
	"net/http"

	"github.com/jmarcosnsf/gobid/internal/jsonutils"
	"github.com/jmarcosnsf/gobid/internal/usecase/user"
)

func (api *Api) handleSignupUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user .CreateUserReq](r)
	if err != nil{
		_ = jsonutils.EncondeJson(w, r, problems, http.StatusUnprocessableEntity)
	}
}

func (api *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	panic("TODO - NOT IMPLEMENTED")
}

func (api *Api) handleLogoutUser(w http.ResponseWriter, r *http.Request) {
	panic("TODO - NOT IMPLEMENTED")
}