package models

import (
	"fmt"
	"time"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

// User represents a user in our system
type User struct {
	Email     string    `dynamodbav:"email"`
	Name      string    `dynamodbav:"name"`
	CreatedAt time.Time `dynamodbav:"created_at"`
}

func (u User) Validate() error {
	if u.Email == "" {
		return &ValidationError{Field: "email", Msg: "email is required"}
	}
	if u.Name == "" {
		return &ValidationError{Field: "name", Msg: "name is required"}
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
		return &ValidationError{Field: "order_id", Msg: "order_id is required"}
	}
	if o.UserEmail == "" {
		return &ValidationError{Field: "user_email", Msg: "user_email is required"}
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

// Validator interface for entities
type Validator interface {
	Validate() error
}
