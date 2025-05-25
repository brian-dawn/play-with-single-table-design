package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
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

// Common errors
var (
	ErrNotFound = errors.New("item not found")
)

// GenericItem makes the Data field type-safe
type GenericItem[T any] struct {
	PK         PrimaryKey `dynamodbav:"PK"`
	SK         SortKey    `dynamodbav:"SK"`
	EntityType string     `dynamodbav:"entity_type"`
	Data       T          `dynamodbav:"data"`
}

// PageToken represents an opaque token for pagination
type PageToken struct {
	PK PrimaryKey `dynamodbav:"PK"`
	SK SortKey    `dynamodbav:"SK"`
}

// QueryOptions contains options for querying items
type QueryOptions struct {
	// Limit is the maximum number of items to return
	Limit int32
	// PageToken is the token for getting the next page
	PageToken *PageToken
}

// QueryResult contains the query results and pagination info
type QueryResult[T any] struct {
	// Items contains the query results
	Items []GenericItem[T]
	// NextPageToken is the token for getting the next page
	// If nil, there are no more pages
	NextPageToken *PageToken
}

// PutItem is a generic function to put any item into DynamoDB
func PutItem[T any](ctx context.Context, s *Store, item GenericItem[T]) error {
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
func GetItem[T any](ctx context.Context, s *Store, pk PrimaryKey, sk SortKey, out *GenericItem[T]) error {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: string(pk)},
			"SK": &types.AttributeValueMemberS{Value: string(sk)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return ErrNotFound
	}

	if err := attributevalue.UnmarshalMap(result.Item, out); err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return nil
}

// Query is a generic function to query items from DynamoDB with pagination support
func Query[T any](ctx context.Context, s *Store, pk PrimaryKey, skPrefix string, opts *QueryOptions) (*QueryResult[T], error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: string(pk)},
			":sk": &types.AttributeValueMemberS{Value: skPrefix},
		},
	}

	// Apply pagination options if provided
	if opts != nil {
		if opts.Limit > 0 {
			queryInput.Limit = aws.Int32(opts.Limit)
		}
		if opts.PageToken != nil {
			exclusiveStartKey, err := attributevalue.MarshalMap(opts.PageToken)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal page token: %w", err)
			}
			queryInput.ExclusiveStartKey = exclusiveStartKey
		}
	}

	result, err := s.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	var items []GenericItem[T]
	for _, item := range result.Items {
		var genericItem GenericItem[T]
		if err := attributevalue.UnmarshalMap(item, &genericItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		items = append(items, genericItem)
	}

	// Handle pagination result
	var nextPageToken *PageToken
	if result.LastEvaluatedKey != nil {
		nextPageToken = &PageToken{}
		if err := attributevalue.UnmarshalMap(result.LastEvaluatedKey, nextPageToken); err != nil {
			return nil, fmt.Errorf("failed to unmarshal last evaluated key: %w", err)
		}
	}

	return &QueryResult[T]{
		Items:         items,
		NextPageToken: nextPageToken,
	}, nil
}
