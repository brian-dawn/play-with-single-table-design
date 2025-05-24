package datastore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"LearnSingleTableDesign/internal/models"
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
func (r *UserRepository) Put(ctx context.Context, user models.User) error {
	item := GenericItem[models.User]{
		PK:         NewUserPK(user.Email),
		SK:         NewUserSK(user.Email),
		EntityType: EntityUser,
		Data:       user,
	}
	return PutItem(ctx, r.store, item)
}

// Get retrieves a user from DynamoDB
func (r *UserRepository) Get(ctx context.Context, email string) (*models.User, error) {
	var item GenericItem[models.User]
	err := GetItem(ctx, r.store, NewUserPK(email), NewUserSK(email), &item)
	if err != nil {
		return nil, err
	}
	return &item.Data, nil
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
func (r *OrderRepository) Put(ctx context.Context, order models.Order) error {
	item := GenericItem[models.Order]{
		PK:         NewUserPK(order.UserEmail),
		SK:         NewOrderSK(order.OrderID),
		EntityType: EntityOrder,
		Data:       order,
	}
	return PutItem(ctx, r.store, item)
}

// OrdersPage represents a page of orders
type OrdersPage struct {
	// Orders in the current page
	Orders []models.Order
	// NextPageToken is the token for getting the next page
	// If nil, there are no more pages
	NextPageToken *PageToken
}

// GetUserOrders retrieves orders for a user from DynamoDB with pagination support
func (r *OrderRepository) GetUserOrders(ctx context.Context, userEmail string, opts *QueryOptions) (*OrdersPage, error) {
	result, err := Query[models.Order](ctx, r.store, NewUserPK(userEmail), "ORDER#", opts)
	if err != nil {
		return nil, err
	}

	orders := make([]models.Order, len(result.Items))
	for i, item := range result.Items {
		orders[i] = item.Data
	}

	return &OrdersPage{
		Orders:        orders,
		NextPageToken: result.NextPageToken,
	}, nil
}
