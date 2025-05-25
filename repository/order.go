package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"LearnSingleTableDesign/models"
)

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

// OrdersPage represents a page of orders
type OrdersPage struct {
	// Orders in the current page
	Orders []models.Order
	// NextPageToken is the token for getting the next page
	// If nil, there are no more pages
	NextPageToken *PageToken
}

// Put stores an order in DynamoDB
func (r *OrderRepository) Put(ctx context.Context, order models.Order) error {
	if err := order.Validate(); err != nil {
		return err
	}
	item := GenericItem[models.Order]{
		PK:         Key.NewUserPK(order.UserEmail),
		SK:         Key.NewOrderSK(order.OrderID),
		EntityType: EntityOrder,
		Data:       order,
	}
	return PutItem(ctx, r.store, item)
}

// GetUserOrders retrieves orders for a user from DynamoDB with pagination support
func (r *OrderRepository) GetUserOrders(ctx context.Context, userEmail string, opts *QueryOptions) (*OrdersPage, error) {
	result, err := Query[models.Order](ctx, r.store, Key.NewUserPK(userEmail), "ORDER#", opts)
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
