package repository

import (
	"LearnSingleTableDesign/models"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type ProductRepository struct {
	store *Store
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
