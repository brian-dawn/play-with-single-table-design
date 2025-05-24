package main

// We can run dynamodb local instance with:
// docker run -d -p 8000:8000 amazon/dynamodb-local

import (
	"context"
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

	// Create multiple orders for the user
	for i := 1; i <= 5; i++ {
		order := models.Order{
			OrderID:   fmt.Sprintf("ORD%d", i),
			UserEmail: user.Email,
			Status:    "PENDING",
			Total:     float64(i) * 10.99,
			CreatedAt: time.Now(),
			Products:  []string{fmt.Sprintf("PROD%d", i)},
		}

		if err := orderRepo.Put(context.TODO(), order); err != nil {
			log.Fatalf("failed to put order: %v", err)
		}
		fmt.Printf("Created order: %s\n", order.OrderID)
	}

	// Demonstrate pagination
	fmt.Println("\nFetching orders with pagination (2 items per page):")
	var pageToken *datastore.PageToken
	pageNum := 1

	for {
		// Get a page of orders
		page, err := orderRepo.GetUserOrders(context.TODO(), user.Email, &datastore.QueryOptions{
			Limit:     2,
			PageToken: pageToken,
		})
		if err != nil {
			log.Fatalf("failed to get user orders: %v", err)
		}

		fmt.Printf("\nPage %d:\n", pageNum)
		for _, order := range page.Orders {
			fmt.Printf("Order: %s, Total: $%.2f\n", order.OrderID, order.Total)
		}

		// If there's no next page token, we've reached the end
		if page.NextPageToken == nil {
			break
		}

		// Set up for next page
		pageToken = page.NextPageToken
		pageNum++
	}
}
