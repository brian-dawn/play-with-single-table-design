package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// UserStore provides type-safe operations for User entities
type UserStore struct {
	store *Store[User]
}

func NewUserStore(client *dynamodb.Client, tableName string) *UserStore {
	return &UserStore{
		store: NewStore[User](client, tableName),
	}
}

func (s *UserStore) PutUser(ctx context.Context, user User) error {
	item := GenericItem[User]{
		PK:         NewUserPK(user.Email),
		SK:         NewUserSK(user.Email),
		EntityType: EntityUser,
		Data:       user,
	}
	return s.store.PutItem(ctx, item)
}

func (s *UserStore) GetUser(ctx context.Context, email string) (*User, error) {
	item, err := s.store.GetItem(ctx, NewUserPK(email), NewUserSK(email))
	if err != nil {
		return nil, err
	}
	return &item.Data, nil
}

// OrderStore provides type-safe operations for Order entities
type OrderStore struct {
	store *Store[Order]
}

func NewOrderStore(client *dynamodb.Client, tableName string) *OrderStore {
	return &OrderStore{
		store: NewStore[Order](client, tableName),
	}
}

func (s *OrderStore) PutOrder(ctx context.Context, order Order) error {
	item := GenericItem[Order]{
		PK:         NewUserPK(order.UserEmail),
		SK:         NewOrderSK(order.OrderID),
		EntityType: EntityOrder,
		Data:       order,
	}
	return s.store.PutItem(ctx, item)
}

func (s *OrderStore) GetUserOrders(ctx context.Context, userEmail string) ([]Order, error) {
	items, err := s.store.Query(ctx, NewUserPK(userEmail), "ORDER#")
	if err != nil {
		return nil, err
	}

	orders := make([]Order, len(items))
	for i, item := range items {
		orders[i] = item.Data
	}
	return orders, nil
}
