package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// IsValid validates if the status is one of the defined constants
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusProcessing, OrderStatusCompleted, OrderStatusCancelled:
		return true
	}
	return false
}

// String converts the OrderStatus to a string
func (s OrderStatus) String() string {
	return string(s)
}

// Value implements the driver.Valuer interface for database operations
func (s OrderStatus) Value() (driver.Value, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid order status: %s", s)
	}
	return string(s), nil
}

// Scan implements the sql.Scanner interface for database operations
func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid order status type: %T", value)
	}
	*s = OrderStatus(str)
	if !s.IsValid() {
		return fmt.Errorf("invalid order status value: %s", str)
	}
	return nil
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
	OrderID   string      `json:"order_id" dynamodbav:"order_id" validate:"required"`
	UserEmail string      `json:"user_email" dynamodbav:"user_email" validate:"required,email"`
	Status    OrderStatus `json:"status" dynamodbav:"status" validate:"required,orderStatus"`
	Total     float64     `json:"total" dynamodbav:"total" validate:"required,gte=0"`
	Products  []string    `json:"products" dynamodbav:"products" validate:"required,min=1,dive,required"`
	CreatedAt time.Time   `json:"created_at" dynamodbav:"created_at"`
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

func init() {
	// Register custom validator for OrderStatus
	validate.RegisterValidation("orderStatus", validateOrderStatus)
}

func validateOrderStatus(fl validator.FieldLevel) bool {
	status, ok := fl.Field().Interface().(OrderStatus)
	if !ok {
		return false
	}
	return status.IsValid()
}
