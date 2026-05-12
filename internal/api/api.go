package api

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/jmarcosnsf/gobid/internal/services"
)

type Api struct {
	Router *chi.Mux
	UserService services.UserService
	ProductSerivce services.ProductService
	BidService services.BidsService
	Sessions *scs.SessionManager
	WsUpgrader websocket.Upgrader
	AuctionLobby services.AuctionLobby
}
