package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmarcosnsf/gobid/internal/jsonutils"
	"github.com/jmarcosnsf/gobid/internal/services"
	"github.com/jmarcosnsf/gobid/internal/usecase/product"
)

func (api *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductReq](r)
	if err != nil{
		jsonutils.EncondeJson(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	userID, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"error":"something went wrong"})
		return
	}

	productId, err := api.ProductSerivce.CreateProduct(r.Context(),userID, data.ProductName, data.Description, data.Baseprice, data.AuctionEnd)
	if err != nil {
		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"error":"failed to create product auction, try again later"})
		return
	}

	ctx, _ := context.WithDeadline(context.Background(), data.AuctionEnd)
	auctionRoom := services.NewAuctionRoom(ctx, productId, api.BidService)

	go auctionRoom.Run()

	api.AuctionLobby.Lock()
	api.AuctionLobby.Rooms[productId] = auctionRoom
	api.AuctionLobby.Unlock()

	jsonutils.EncondeJson(w, r, http.StatusCreated, map[string]any{
		"message":"auction has started with success",
		"product_id": productId,
	})

}
