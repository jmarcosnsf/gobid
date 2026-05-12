package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	// Requests
	PlaceBid MessageKind = iota
	// Ok/Success
	SucessfullyPlacedBid
	// Errors
	FailedToPlaceBid
	InvalidJSON
	// Info
	NewBidPlaced
	AuctionFinished
)

type Message struct {
	Message string      `json:"message,omitempty"`
	Amount  float64     `json:"amount,omitempty"`
	Kind    MessageKind `json:"kind"`
	UserID  uuid.UUID   `json:"user_id,omitempty"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	Id         uuid.UUID
	Context    context.Context
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Clients    map[uuid.UUID]*Client

	BidsService *BidsService
}

func (r *AuctionRoom) registerClient(c *Client) {
	slog.Info("New user connected", "Client: ", c)
	r.Clients[c.UserId] = c
}

func (r *AuctionRoom) unregisterClient(c *Client) {
	slog.Info("User disconnected", "Client: ", c)
	delete(r.Clients, c.UserId)
}

func (r *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message recieved", "RoomID", r.Id, "message", m.Message, "user_id", m.UserID)
	switch m.Kind {
	case (PlaceBid):
		bid, err := r.BidsService.PlaceBid(r.Context, r.Id, m.UserID, m.Amount)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				if client, ok := r.Clients[m.UserID]; ok {
					client.Send <- Message{Kind: FailedToPlaceBid, Message: ErrBidIsTooLow.Error(), UserID: m.UserID}
				}
				return
			}

			slog.Error("failed to place bid", "error", err)
			if client, ok := r.Clients[m.UserID]; ok {
				client.Send <- Message{Kind: FailedToPlaceBid, Message: "failed to place bid", UserID: m.UserID}
			}
			return
		}

		if client, ok := r.Clients[m.UserID]; ok {
			client.Send <- Message{Kind: SucessfullyPlacedBid, Message: "Your bid was sucessfully placed", UserID: m.UserID}
		}

		for id, client := range r.Clients {
			newBidMessage := Message{Kind: NewBidPlaced, Message: "A new bid has been placed", Amount: bid.BidAmount, UserID: m.UserID}
			if id == m.UserID {
				continue
			}
			client.Send <- newBidMessage
		}
	case InvalidJSON:
		client, ok := r.Clients[m.UserID]
		if !ok {
			slog.Info("Client not found in hashmap", "user_id", m.UserID)
			return
		}
		client.Send <- m
	}

}
func (r *AuctionRoom) Run() {
	slog.Info("Auction has begun", "auctionID", r.Id)
	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
	}()

	for {
		select {
		case client := <-r.Register:
			r.registerClient(client)
		case client := <-r.Unregister:
			r.unregisterClient(client)
		case message := <-r.Broadcast:
			r.broadcastMessage(message)
		case <-r.Context.Done():
			slog.Info("Auction has ended.", "auctionID", r.Id)
			for _, client := range r.Clients {
				client.Send <- Message{Kind: AuctionFinished, Message: "auction has been finished"}
			}
			return
		}
	}
}
func NewAuctionRoom(ctx context.Context, id uuid.UUID, bidsService BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:          id,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]*Client),
		Context:     ctx,
		BidsService: &bidsService,
	}
}

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserId uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserId: userId,
	}
}

const (
	maxMessageSize = 512
	readDeadLine   = 60 * time.Second
	writeWait      = 10 * time.Second
	pingPeriod     = (readDeadLine * 9) / 10
)

func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(readDeadLine))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(readDeadLine))
		return nil
	})

	for {
		var m Message
		m.UserID = c.UserId
		if err := c.Conn.ReadJSON(&m); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected close error", "error", err)
				return
			}

			c.Room.Broadcast <- Message{
				Kind:    InvalidJSON,
				Message: "this message should be a valid json",
				UserID:  m.UserID,
			}
			continue
		}

		c.Room.Broadcast <- m
	}
}

func (c *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Message: "closing websocket connection",
				})
				return
			}

			if message.Kind == AuctionFinished {
				close(c.Send)
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.Conn.WriteJSON(message); err != nil {
				c.Room.Unregister <- c
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Unexpected write error", "error", err)
				return
			}
		}
	}
}
