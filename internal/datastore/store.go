package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Common errors
var (
	ErrNotFound = errors.New("item not found")
)

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

// Query is a generic function to query items from DynamoDB
func Query[T any](ctx context.Context, s *Store, pk PrimaryKey, skPrefix string) ([]GenericItem[T], error) {
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
	for _, item := range result.Items {
		var genericItem GenericItem[T]
		if err := attributevalue.UnmarshalMap(item, &genericItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		items = append(items, genericItem)
	}

	return items, nil
}
