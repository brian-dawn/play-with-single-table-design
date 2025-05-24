package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

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
