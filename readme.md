# Learn Single Table Design with DynamoDB

This project demonstrates single table design patterns with DynamoDB using Go. It shows how to store and query multiple entity types (Users, Orders, Products) in a single DynamoDB table.

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose

## Quick Start

1. Start DynamoDB Local and Admin UI:
```bash
docker-compose up -d
```

2. Run the example:
```bash
go run main.go
```

Services:
- DynamoDB Local: http://localhost:8000
- DynamoDB Admin UI: http://localhost:8001

## Data Model

### Key Patterns
- Users: 
  - PK: `USER#<email>`
  - SK: `PROFILE#<email>`
- Orders:
  - PK: `USER#<email>`
  - SK: `ORDER#<orderID>`
- Products:
  - PK: `PRODUCT#<sku>`
  - SK: `PRODUCT#<sku>`

### Example Usage

```go
// Create a store instance
store := NewStore(dynamoClient, "AppTable")

// Create a user
user := User{
    Email:     "john@example.com",
    Name:      "John Doe",
    CreatedAt: time.Now(),
}
err := store.PutUser(ctx, user)

// Get a user
retrievedUser, err := store.GetUser(ctx, "john@example.com")

// Create an order
order := Order{
    OrderID:   "ORD123",
    UserEmail: "john@example.com",
    Status:    "PENDING",
    Total:     99.99,
    CreatedAt: time.Now(),
    Products:  []string{"PROD1", "PROD2"},
}
err = store.PutOrder(ctx, order)

// Get all orders for a user
orders, err := store.GetUserOrders(ctx, "john@example.com")
```

## Features

- ✅ Single table design pattern
- ✅ Type-safe Go structs with DynamoDB mappings
- ✅ Automatic table creation
- ✅ Clean helper functions for common operations
- ✅ Local development environment with admin UI