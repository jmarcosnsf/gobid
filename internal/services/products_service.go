package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmarcosnsf/gobid/internal/store/pgstore"
)

type ProductService struct {
	pool *pgxpool.Pool
	queries *pgstore.Queries
}

func NewProductService(pool *pgxpool.Pool) ProductService {
	return ProductService{
		pool: pool,
		queries: pgstore.New(pool),
	}
}

func (ps *ProductService) CreateProduct(ctx context.Context, seller_id uuid.UUID, productname, description string, baseprice float64, auctioneEnd time.Time) (uuid.UUID, error){

	id, err := ps.queries.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID: seller_id,
		ProductName: productname,
		Description: description,
		Baseprice: baseprice,
		AuctionEnd: auctioneEnd,
	})
	if err != nil{
		return uuid.UUID{}, err
	}

	return id, nil
}

var ErrProductNotFound = errors.New("product not found")

func (ps *ProductService) GetProductById(ctx context.Context, productId uuid.UUID) (pgstore.Product, error) {
	product, err := ps.queries.GetProductById(ctx, productId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return pgstore.Product{}, ErrProductNotFound
		}
		return pgstore.Product{}, err
	}

	return product, nil
}