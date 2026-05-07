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
	if err != nil {
		if errors.Is(err, services.ErrDuplicateEmailOrUsername) {
			_ = jsonutils.EncondeJson(w, r, http.StatusUnprocessableEntity, map[string]any{"error": "username or email already exists"})
			return
		}
	}

	_ = jsonutils.EncondeJson(w, r, http.StatusAccepted, map[string]any{"user_id": id})

}

func (api *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.LoginUserRequest](r)
	if err != nil {
		jsonutils.EncondeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	id, err := api.UserService.AuthenticateUser(r.Context(), data.Email, data.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			jsonutils.EncondeJson(w, r, http.StatusBadRequest, map[string]any{"error": "invalid email or password"})
			return
		}

		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"error": "something went wrong"})
		return
	}

	if err := api.Sessions.RenewToken(r.Context()); err != nil {
		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"error": "something went wrong"})
		return
	}

	api.Sessions.Put(r.Context(), "AuthenticatedUserId", id)

	jsonutils.EncondeJson(w, r, http.StatusOK, map[string]any{"message": "sucessfully logged in"})
}

func (api *Api) handleLogoutUser(w http.ResponseWriter, r *http.Request) {
	if err := api.Sessions.RenewToken(r.Context()); err != nil {
		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"error": "something went wrong"})
		return
	}

	api.Sessions.Remove(r.Context(), "AuthenticatedUserId")

	jsonutils.EncondeJson(w, r, http.StatusOK, map[string]any{"message": "sucessfully logged out"})
}
