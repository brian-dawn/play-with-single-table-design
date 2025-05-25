package models

import (
	"errors"
	"time"
)

// User represents a user in the system
type User struct {
	Email     string    `json:"email" dynamodbav:"email"`
	Name      string    `json:"name" dynamodbav:"name"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
}

// Validate validates the user fields
func (u User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// Order represents an order in the system
type Order struct {
	OrderID   string    `json:"order_id" dynamodbav:"order_id"`
	UserEmail string    `json:"user_email" dynamodbav:"user_email"`
	Status    string    `json:"status" dynamodbav:"status"`
	Total     float64   `json:"total" dynamodbav:"total"`
	Products  []string  `json:"products" dynamodbav:"products"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
}

// Validate validates the order fields
func (o Order) Validate() error {
	if o.OrderID == "" {
		return errors.New("order_id is required")
	}
	if o.UserEmail == "" {
		return errors.New("user_email is required")
	}
	if o.Status == "" {
		return errors.New("status is required")
	}
	if len(o.Products) == 0 {
		return errors.New("at least one product is required")
	}
	return nil
}

type Product struct {
	ProductID string  `json:"product_id" dynamodbav:"product_id"`
	Category  string  `json:"category" dynamodbav:"category"`
	Name      string  `json:"name" dynamodbav:"name"`
	Price     float64 `json:"price" dynamodbav:"price"`
	Stock     int     `json:"stock" dynamodbav:"stock"`
}

func (p Product) Validate() error {
	if p.ProductID == "" {
		return errors.New("product_id is required")
	}
	if p.Category == "" {
		return errors.New("category is required")
	}
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.Price <= 0 {
		return errors.New("price is required")
	}
	if p.Stock < 0 {
		return errors.New("stock can't be less than 0")
	}
	return nil
}
