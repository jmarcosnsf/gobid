package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmarcosnsf/gobid/internal/jsonutils"
	"github.com/jmarcosnsf/gobid/internal/services"
)

func (api *Api) handleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	rawProductId := chi.URLParam(r, "product_id")

	productId, err := uuid.Parse(rawProductId)
	if err != nil {
		jsonutils.EncondeJson(w, r, http.StatusBadRequest, map[string]any{"message": "invalid product id - must be a valid uuid"})
		return
	}

	_, err = api.ProductSerivce.GetProductById(r.Context(), productId)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncondeJson(w, r, http.StatusNotFound, map[string]any{"message": "no product with given id"})
			return
		}
		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"message": "something went wrong"})
		return
	}

	userId, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"message": "something went wrong"})
		return
	}

	api.AuctionLobby.Lock()
	room, ok := api.AuctionLobby.Rooms[productId]
	api.AuctionLobby.Unlock()

	if !ok {
		jsonutils.EncondeJson(w, r, http.StatusBadRequest, map[string]any{"message": "the auction has ended"})
		return
	}

	conn, err := api.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"message": "could not upgrade connection to a websocket protocol"})
		return
	}

	client := services.NewClient(room, conn, userId)
	
	room.Register <- client

	// go client.ReadEventLoop()
	// go client.WriteEventLoop()
	for {
		
	}
	
	
}
