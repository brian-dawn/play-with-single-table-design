package datastore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// UserRepository handles User entity operations
type UserRepository struct {
	store *Store
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(client *dynamodb.Client, tableName string) *UserRepository {
	return &UserRepository{
		store: NewStore(client, tableName),
	}
}

// Put stores a user in DynamoDB
func (r *UserRepository) Put(ctx context.Context, user interface{}) error {
	item := GenericItem[interface{}]{
		PK:         NewUserPK(user.(struct{ Email string }).Email),
		SK:         NewUserSK(user.(struct{ Email string }).Email),
		EntityType: EntityUser,
		Data:       user,
	}
	return r.store.PutItem(ctx, item)
}

// Get retrieves a user from DynamoDB
func (r *UserRepository) Get(ctx context.Context, email string) (interface{}, error) {
	var item GenericItem[interface{}]
	err := r.store.GetItem(ctx, NewUserPK(email), NewUserSK(email), &item)
	if err != nil {
		return nil, err
	}
	return item.Data, nil
}

// OrderRepository handles Order entity operations
type OrderRepository struct {
	store *Store
}

// NewOrderRepository creates a new OrderRepository
func NewOrderRepository(client *dynamodb.Client, tableName string) *OrderRepository {
	return &OrderRepository{
		store: NewStore(client, tableName),
	}
}

// Put stores an order in DynamoDB
func (r *OrderRepository) Put(ctx context.Context, order interface{}) error {
	o := order.(struct {
		OrderID   string
		UserEmail string
	})
	item := GenericItem[interface{}]{
		PK:         NewUserPK(o.UserEmail),
		SK:         NewOrderSK(o.OrderID),
		EntityType: EntityOrder,
		Data:       order,
	}
	return r.store.PutItem(ctx, item)
}

// GetUserOrders retrieves all orders for a user from DynamoDB
func (r *OrderRepository) GetUserOrders(ctx context.Context, userEmail string) ([]interface{}, error) {
	items, err := r.store.Query(ctx, NewUserPK(userEmail), "ORDER#")
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for _, item := range items {
		var genericItem GenericItem[interface{}]
		if err := attributevalue.UnmarshalMap(item, &genericItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		results = append(results, genericItem.Data)
	}
	return results, nil
}
