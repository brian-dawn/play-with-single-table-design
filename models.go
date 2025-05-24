package main

import (
	"fmt"
	"time"
)

// User represents a user in our system
type User struct {
	Email     string    `dynamodbav:"email"`
	Name      string    `dynamodbav:"name"`
	CreatedAt time.Time `dynamodbav:"created_at"`
}

func (u User) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidData)
	}
	if u.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidData)
	}
	return nil
}

// Order represents an order in our system
type Order struct {
	OrderID   string    `dynamodbav:"order_id"`
	UserEmail string    `dynamodbav:"user_email"`
	Status    string    `dynamodbav:"status"`
	Total     float64   `dynamodbav:"total"`
	CreatedAt time.Time `dynamodbav:"created_at"`
	Products  []string  `dynamodbav:"products"`
}

func (o Order) Validate() error {
	if o.OrderID == "" {
		return fmt.Errorf("%w: order_id is required", ErrInvalidData)
	}
	if o.UserEmail == "" {
		return fmt.Errorf("%w: user_email is required", ErrInvalidData)
	}
	return nil
}

// Product represents a product in our system
type Product struct {
	SKU       string    `dynamodbav:"sku"`
	Name      string    `dynamodbav:"name"`
	Price     float64   `dynamodbav:"price"`
	CreatedAt time.Time `dynamodbav:"created_at"`
}
