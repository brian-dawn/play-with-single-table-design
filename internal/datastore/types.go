package datastore

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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

// GenericItem makes the Data field type-safe
type GenericItem[T any] struct {
	PK         PrimaryKey `dynamodbav:"PK"`
	SK         SortKey    `dynamodbav:"SK"`
	EntityType string     `dynamodbav:"entity_type"`
	Data       T          `dynamodbav:"data"`
}

// Store represents a DynamoDB store
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
