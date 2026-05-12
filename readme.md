# GoBid

Real-time auction API in Go. Users list products for auction and bid on them in real time over WebSockets.

## Stack

- **Go 1.25**
- **Chi** — HTTP router
- **PostgreSQL 16** with **pgx/pgxpool**
- **sqlc** — type-safe SQL code generation
- **tern** — database migrations
- **alexedwards/scs** — server-side sessions (Postgres-backed)
- **gorilla/websocket** — real-time bidding
- **bcrypt** — password hashing
- **Air** — live reload during development

## How it works

When a product is created, the server spawns an `AuctionRoom` (goroutine) bound to the auction's deadline via `context.WithDeadline`. Clients connect through `/ws/subscribe/{product_id}` and exchange typed JSON messages. The `AuctionLobby` keeps track of all active rooms. Auth is session-based, stored in Postgres via SCS. When the deadline hits, the room broadcasts `AuctionFinished` and shuts down.

## Running

Requirements: Go 1.25+, Docker, [tern](https://github.com/jackc/tern).

```bash
cp .env.example .env       # fill in DB credentials
docker compose up -d       # start Postgres
go run ./cmd/terndotenv    # run migrations
go run ./cmd/api           # API on :3080
```

## API

| Method | Path                                         | Auth |
| ------ | -------------------------------------------- | ---- |
| POST   | `/api/v1/users/signup`                       |      |
| POST   | `/api/v1/users/login`                        |      |
| POST   | `/api/v1/users/logout`                       | ✅   |
| POST   | `/api/v1/products`                           | ✅   |
| GET    | `/api/v1/products/ws/subscribe/{product_id}` | ✅   |

Example — create a product (starts an auction):

```http
POST /api/v1/products
Content-Type: application/json

{
  "product_name": "Vintage Camera",
  "description": "Leica M3 in great condition",
  "baseprice": 500.00,
  "auction_end": "2026-06-01T18:00:00Z"
}
```

### WebSocket protocol

Messages are JSON with a `kind` discriminator:

| Kind | Direction       | Purpose                 |
| ---- | --------------- | ----------------------- |
| 0    | client → server | `PlaceBid`              |
| 1    | server → bidder | `SuccessfullyPlacedBid` |
| 2    | server → bidder | `FailedToPlaceBid`      |
| 3    | server → bidder | `InvalidJSON`           |
| 4    | server → room   | `NewBidPlaced`          |
| 5    | server → room   | `AuctionFinished`       |

Place a bid:

```json
{ "kind": 0, "amount": 750.0 }
```

## Roadmap

- Prevent sellers from bidding on their own products
- Product listing endpoint with pagination
- Endpoint to fetch an auction's winner (currently only available via WebSocket)
- Broadcast the winner over WebSocket when the auction closes
- CSRF protection (scaffolded with `gorilla/csrf`, not yet wired)