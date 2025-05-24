package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// EnsureTableExists creates the DynamoDB table if it doesn't already exist
func EnsureTableExists(ctx context.Context, svc *dynamodb.Client, tableName string) error {
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
