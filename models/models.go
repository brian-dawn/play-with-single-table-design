package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// User represents a user in the system
type User struct {
	Email     string    `json:"email" dynamodbav:"email" validate:"required,email"`
	Name      string    `json:"name" dynamodbav:"name" validate:"required"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
}

// Validate validates the user fields
func (u User) Validate() error {
	return validate.Struct(u)
}

// Order represents an order in the system
type Order struct {
	OrderID   string    `json:"order_id" dynamodbav:"order_id" validate:"required"`
	UserEmail string    `json:"user_email" dynamodbav:"user_email" validate:"required,email"`
	Status    string    `json:"status" dynamodbav:"status" validate:"required,oneof=pending processing completed cancelled"`
	Total     float64   `json:"total" dynamodbav:"total" validate:"required,gte=0"`
	Products  []string  `json:"products" dynamodbav:"products" validate:"required,min=1,dive,required"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
}

// Validate validates the order fields
func (o Order) Validate() error {
	return validate.Struct(o)
}

type Product struct {
	ProductID string    `json:"product_id" dynamodbav:"product_id" validate:"required"`
	Category  string    `json:"category" dynamodbav:"category" validate:"required"`
	Name      string    `json:"name" dynamodbav:"name" validate:"required"`
	Price     float64   `json:"price" dynamodbav:"price" validate:"required,gt=0"`
	Stock     int       `json:"stock" dynamodbav:"stock" validate:"gte=0"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
}

func (p Product) Validate() error {
	return validate.Struct(p)
}
