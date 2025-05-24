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

// Custom key types for type safety
type PrimaryKey string
type SortKey string

// Key constructors
func NewUserPK(email string) PrimaryKey {
	return PrimaryKey(fmt.Sprintf("USER#%s", email))
}

func NewUserSK(email string) SortKey {
	return SortKey(fmt.Sprintf("PROFILE#%s", email))
}

func NewOrderSK(orderID string) SortKey {
	return SortKey(fmt.Sprintf("ORDER#%s", orderID))
}

// Error types
var (
	ErrNotFound    = errors.New("item not found")
	ErrInvalidData = errors.New("invalid data")
)

// Validator interface for entities
type Validator interface {
	Validate() error
}

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

// GenericItem makes the Data field type-safe
type GenericItem[T any] struct {
	PK         PrimaryKey `dynamodbav:"PK"`
	SK         SortKey    `dynamodbav:"SK"`
	EntityType string     `dynamodbav:"entity_type"`
	Data       T          `dynamodbav:"data"`
}

// Store handles DynamoDB operations with type safety
type Store[T Validator] struct {
	client    *dynamodb.Client
	tableName string
}

// NewStore creates a new Store instance
func NewStore[T Validator](client *dynamodb.Client, tableName string) *Store[T] {
	return &Store[T]{
		client:    client,
		tableName: tableName,
	}
}

// PutItem is a generic function to put any item into DynamoDB
func (s *Store[T]) PutItem(ctx context.Context, item GenericItem[T]) error {
	if err := item.Data.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

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
func (s *Store[T]) GetItem(ctx context.Context, pk PrimaryKey, sk SortKey) (*GenericItem[T], error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: string(pk)},
			"SK": &types.AttributeValueMemberS{Value: string(sk)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return nil, ErrNotFound
	}

	var item GenericItem[T]
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &item, nil
}

// Query is a generic function to query items from DynamoDB
func (s *Store[T]) Query(ctx context.Context, pk PrimaryKey, skPrefix string) ([]GenericItem[T], error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: string(pk)},
			":sk": &types.AttributeValueMemberS{Value: skPrefix},
		},
	}

	result, err := s.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	var items []GenericItem[T]
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal items: %w", err)
	}

	return items, nil
}

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

	// Create store instances
	tableName := "AppTable"
	userStore := NewUserStore(client, tableName)
	orderStore := NewOrderStore(client, tableName)

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
	if err := userStore.PutUser(context.TODO(), user); err != nil {
		log.Fatalf("failed to put user: %v", err)
	}
	fmt.Println("Successfully created user:", user.Email)

	// Get user from DynamoDB
	retrievedUser, err := userStore.GetUser(context.TODO(), user.Email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			fmt.Println("User not found")
		} else {
			log.Fatalf("failed to get user: %v", err)
		}
	} else {
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
	if err := orderStore.PutOrder(context.TODO(), order); err != nil {
		log.Fatalf("failed to put order: %v", err)
	}
	fmt.Println("Successfully created order:", order.OrderID)

	// Get all orders for the user
	orders, err := orderStore.GetUserOrders(context.TODO(), user.Email)
	if err != nil {
		log.Fatalf("failed to get user orders: %v", err)
	}

	fmt.Printf("Found %d orders for user %s\n", len(orders), user.Email)
	for _, order := range orders {
		fmt.Printf("Order: %+v\n", order)
	}
}
