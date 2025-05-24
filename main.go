package main

// We can run dynamodb local instance with:
// docker run -d -p 8000:8000 amazon/dynamodb-local

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"LearnSingleTableDesign/internal/datastore"
	"LearnSingleTableDesign/internal/db"
	"LearnSingleTableDesign/internal/models"
)

func main() {
	// Create custom resolver to point to local DynamoDB
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost:8000",
			SigningRegion: "us-east-1",
		}, nil
	})

	// Configure AWS SDK with local endpoint
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; DO NOT use in production",
			},
		}),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	// Create repositories
	tableName := "AppTable"
	userRepo := datastore.NewUserRepository(client, tableName)
	orderRepo := datastore.NewOrderRepository(client, tableName)

	// Ensure the table exists before proceeding
	if err := db.EnsureTableExists(context.TODO(), client, tableName); err != nil {
		log.Fatalf("failed to ensure table exists: %v", err)
	}

	// Example: Create a new user
	user := models.User{
		Email:     "john@example.com",
		Name:      "John Doe",
		CreatedAt: time.Now(),
	}

	// Put user in DynamoDB
	if err := userRepo.Put(context.TODO(), user); err != nil {
		log.Fatalf("failed to put user: %v", err)
	}
	fmt.Println("Successfully created user:", user.Email)

	// Get user from DynamoDB
	retrievedUser, err := userRepo.Get(context.TODO(), user.Email)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			fmt.Println("User not found")
		} else {
			log.Fatalf("failed to get user: %v", err)
		}
	} else {
		fmt.Printf("Retrieved user: %+v\n", retrievedUser)
	}

	// Example: Create an order for the user
	order := models.Order{
		OrderID:   "ORD123",
		UserEmail: user.Email,
		Status:    "PENDING",
		Total:     99.99,
		CreatedAt: time.Now(),
		Products:  []string{"PROD1", "PROD2"},
	}

	// Put order in DynamoDB
	if err := orderRepo.Put(context.TODO(), order); err != nil {
		log.Fatalf("failed to put order: %v", err)
	}
	fmt.Println("Successfully created order:", order.OrderID)

	// Get all orders for the user
	orders, err := orderRepo.GetUserOrders(context.TODO(), user.Email)
	if err != nil {
		log.Fatalf("failed to get user orders: %v", err)
	}

	fmt.Printf("Found %d orders for user %s\n", len(orders), user.Email)
	for _, order := range orders {
		fmt.Printf("Order: %+v\n", order)
	}
}
