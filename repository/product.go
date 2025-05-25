package repository

import (
	"LearnSingleTableDesign/models"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type ProductRepository struct {
	store *Store
}

type ProductsPage struct {
	Products      []models.Product
	NextPageToken *PageToken
}

func NewProductRepository(client *dynamodb.Client, tableName string) *ProductRepository {
	return &ProductRepository{
		store: NewStore(client, tableName),
	}
}

func (r *ProductRepository) Put(ctx context.Context, product models.Product) error {
	if err := product.Validate(); err != nil {
		return err
	}
	item := GenericItem[models.Product]{
		PK:         Key.ProductPK(),
		SK:         Key.ProductSK(product.ProductID),
		EntityType: EntityProduct,
		Data:       product,
	}
	return PutItem(ctx, r.store, item)
}

func (r *ProductRepository) Get(ctx context.Context, productID string) (*models.Product, error) {
	var item GenericItem[models.Product]
	err := GetItem(ctx, r.store, Key.ProductPK(), Key.ProductSK(productID), &item)
	if err != nil {
		return nil, err
	}
	return &item.Data, nil
}

func (r *ProductRepository) All(ctx context.Context, opts *QueryOptions) (*ProductsPage, error) {
	result, err := Query[models.Product](ctx, r.store, Key.ProductPK(), "PRODUCT#", opts)
	if err != nil {
		return nil, err
	}

	products := make([]models.Product, len(result.Items))
	for i, item := range result.Items {
		products[i] = item.Data
	}

	return &ProductsPage{
		Products:      products,
		NextPageToken: result.NextPageToken,
	}, nil
}
