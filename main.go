package main

// We can run dynamodb local instance with:

// docker run -d -p 8000:8000 amazon/dynamodb-local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Entity types for our single table design
const (
	EntityUser    = "USER"
	EntityOrder   = "ORDER"
	EntityProduct = "PRODUCT"
)

// User represents a user in our system
type User struct {
	Email     string    `dynamodbav:"email"`
	Name      string    `dynamodbav:"name"`
	CreatedAt time.Time `dynamodbav:"created_at"`
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

// Product represents a product in our system
type Product struct {
	SKU       string    `dynamodbav:"sku"`
	Name      string    `dynamodbav:"name"`
	Price     float64   `dynamodbav:"price"`
	CreatedAt time.Time `dynamodbav:"created_at"`
}

// Item represents our DynamoDB item structure
type Item struct {
	PK         string      `dynamodbav:"PK"`
	SK         string      `dynamodbav:"SK"`
	EntityType string      `dynamodbav:"entity_type"`
	Data       interface{} `dynamodbav:"data"`
}

// Store handles DynamoDB operations
type Store struct {
	client    *dynamodb.Client
	tableName string
}

// NewStore creates a new Store instance
func NewStore(client *dynamodb.Client, tableName string) *Store {
	return &Store{
		client:    client,
		tableName: tableName,
	}
}

// PutItem is a generic function to put any item into DynamoDB
func (s *Store) PutItem(ctx context.Context, item Item) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      av,
	})
	return err
}

// GetItem is a generic function to get any item from DynamoDB
func (s *Store) GetItem(ctx context.Context, pk, sk string) (*Item, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var item Item
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &item, nil
}

// Query is a generic function to query items from DynamoDB
func (s *Store) Query(ctx context.Context, pk, skPrefix string) ([]Item, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
			":sk": &types.AttributeValueMemberS{Value: skPrefix},
		},
	}

	result, err := s.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	var items []Item
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal items: %w", err)
	}

	return items, nil
}

// UnmarshalData unmarshals the Data field of an Item into the provided interface
func UnmarshalData(item *Item, v interface{}) error {
	bytes, err := json.Marshal(item.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	if err := json.Unmarshal(bytes, v); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// User-specific helper functions
func (s *Store) PutUser(ctx context.Context, user User) error {
	item := Item{
		PK:         fmt.Sprintf("USER#%s", user.Email),
		SK:         fmt.Sprintf("PROFILE#%s", user.Email),
		EntityType: EntityUser,
		Data:       user,
	}
	return s.PutItem(ctx, item)
}

func (s *Store) GetUser(ctx context.Context, email string) (*User, error) {
	item, err := s.GetItem(ctx, fmt.Sprintf("USER#%s", email), fmt.Sprintf("PROFILE#%s", email))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	var user User
	if err := UnmarshalData(item, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Order-specific helper functions
func (s *Store) PutOrder(ctx context.Context, order Order) error {
	item := Item{
		PK:         fmt.Sprintf("USER#%s", order.UserEmail),
		SK:         fmt.Sprintf("ORDER#%s", order.OrderID),
		EntityType: EntityOrder,
		Data:       order,
	}
	return s.PutItem(ctx, item)
}

func (s *Store) GetUserOrders(ctx context.Context, userEmail string) ([]Order, error) {
	items, err := s.Query(ctx, fmt.Sprintf("USER#%s", userEmail), "ORDER#")
	if err != nil {
		return nil, err
	}

	orders := make([]Order, 0, len(items))
	for _, item := range items {
		var order Order
		if err := UnmarshalData(&item, &order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// ensureTableExists creates the DynamoDB table if it doesn't already exist
func ensureTableExists(ctx context.Context, svc *dynamodb.Client, tableName string) error {
	// Check if table exists
	_, err := svc.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		// If error is table doesn't exist, create it
		var notFoundEx *types.ResourceNotFoundException
		if ok := errors.As(err, &notFoundEx); ok {
			fmt.Printf("Table %s does not exist, creating...\n", tableName)

			_, err = svc.CreateTable(ctx, &dynamodb.CreateTableInput{
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("PK"),
						AttributeType: types.ScalarAttributeTypeS,
					},
					{
						AttributeName: aws.String("SK"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				TableName:   aws.String(tableName),
				BillingMode: types.BillingModePayPerRequest,
			})
			if err != nil {
				return fmt.Errorf("failed to create table: %w", err)
			}

			// Wait for table to be active
			waiter := dynamodb.NewTableExistsWaiter(svc)
			err = waiter.Wait(ctx,
				&dynamodb.DescribeTableInput{
					TableName: aws.String(tableName),
				},
				2*time.Minute,
			)
			if err != nil {
				return fmt.Errorf("timeout waiting for table creation: %w", err)
			}

			fmt.Printf("Table %s created successfully\n", tableName)
		} else {
			return fmt.Errorf("error checking if table exists: %w", err)
		}
	}

	return nil
}

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

	// Create store instance
	tableName := "AppTable"
	store := NewStore(client, tableName)

	// Ensure the table exists before proceeding
	if err := ensureTableExists(context.TODO(), client, tableName); err != nil {
		log.Fatalf("failed to ensure table exists: %v", err)
	}

	// Example: Create a new user
	user := User{
		Email:     "john@example.com",
		Name:      "John Doe",
		CreatedAt: time.Now(),
	}

	// Put user in DynamoDB
	if err := store.PutUser(context.TODO(), user); err != nil {
		log.Fatalf("failed to put user: %v", err)
	}
	fmt.Println("Successfully created user:", user.Email)

	// Get user from DynamoDB
	retrievedUser, err := store.GetUser(context.TODO(), user.Email)
	if err != nil {
		log.Fatalf("failed to get user: %v", err)
	}
	if retrievedUser != nil {
		fmt.Printf("Retrieved user: %+v\n", retrievedUser)
	}

	// Example: Create an order for the user
	order := Order{
		OrderID:   "ORD123",
		UserEmail: user.Email,
		Status:    "PENDING",
		Total:     99.99,
		CreatedAt: time.Now(),
		Products:  []string{"PROD1", "PROD2"},
	}

	// Put order in DynamoDB
	if err := store.PutOrder(context.TODO(), order); err != nil {
		log.Fatalf("failed to put order: %v", err)
	}
	fmt.Println("Successfully created order:", order.OrderID)

	// Get all orders for the user
	orders, err := store.GetUserOrders(context.TODO(), user.Email)
	if err != nil {
		log.Fatalf("failed to get user orders: %v", err)
	}

	fmt.Printf("Found %d orders for user %s\n", len(orders), user.Email)
	for _, order := range orders {
		fmt.Printf("Order: %+v\n", order)
	}
}
