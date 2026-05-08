package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jmarcosnsf/gobid/internal/jsonutils"
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

	id, err := api.ProductSerivce.CreateProduct(r.Context(),userID, data.ProductName, data.Description, data.Baseprice, data.AuctionEnd)
	if err != nil {
		jsonutils.EncondeJson(w, r, http.StatusInternalServerError, map[string]any{"error":"failed to create product auction, try again later"})
		return
	}

	jsonutils.EncondeJson(w, r, http.StatusCreated, map[string]any{
		"message":"product created with success",
		"product_id": id,
	})

}
