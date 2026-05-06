package api

import (
	"errors"
	"net/http"

	"github.com/jmarcosnsf/gobid/internal/jsonutils"
	"github.com/jmarcosnsf/gobid/internal/services"
	"github.com/jmarcosnsf/gobid/internal/usecase/user"
)

func (api *Api) handleSignupUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.CreateUserReq](r)
	if err != nil {
		_ = jsonutils.EncondeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	id, err := api.UserService.CreateUser(r.Context(),
		data.Username,
		data.Email,
		data.Password,
		data.Bio,
	)
	if err != nil{
		if errors.Is(err, services.ErrDuplicateEmailOrPassword){
			_ = jsonutils.EncondeJson(w, r, http.StatusUnprocessableEntity, map[string]any{"error": "username or email already exists"})
			return
		}
	}

	_ = jsonutils.EncondeJson(w, r, http.StatusAccepted, map[string]any{"user_id": id})

}

func (api *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	panic("TODO - NOT IMPLEMENTED")
}

func (api *Api) handleLogoutUser(w http.ResponseWriter, r *http.Request) {
	panic("TODO - NOT IMPLEMENTED")
}
