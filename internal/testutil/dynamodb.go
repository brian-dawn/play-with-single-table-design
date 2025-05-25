package testutil

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"LearnSingleTableDesign/internal/db"
)

// sanitizeTableName ensures the table name is valid for DynamoDB
func sanitizeTableName(name string) string {
	// Replace invalid characters with underscores
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			return r
		}
		return '_'
	}, name)

	// Ensure name starts with a letter or number
	if len(name) > 0 && !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z') || (name[0] >= '0' && name[0] <= '9')) {
		name = "t_" + name
	}

	// Ensure minimum length
	if len(name) < 3 {
		name = name + "_table"
	}

	return name
}

// CreateTestClient creates a DynamoDB client configured to use the local DynamoDB instance
func CreateTestClient(t *testing.T) *dynamodb.Client {
	t.Helper()

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost:8000",
			SigningRegion: "us-east-1",
		}, nil
	})

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
		t.Fatalf("unable to load SDK config: %v", err)
	}

	return dynamodb.NewFromConfig(cfg)
}

// SetupTestTable creates a new table with a unique name for testing
func SetupTestTable(t *testing.T, client *dynamodb.Client) string {
	t.Helper()

	// Create a unique table name for this test
	tableName := sanitizeTableName(fmt.Sprintf("test_%s", t.Name()))

	// Ensure any existing table is deleted
	_, err := client.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		// Ignore if table doesn't exist
		var notFoundErr *types.ResourceNotFoundException
		if !errors.As(err, &notFoundErr) {
			t.Logf("error deleting table (ignored): %v", err)
		}
	}

	// Create the table
	err = db.EnsureTableExists(context.TODO(), client, tableName)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Return the table name so it can be used in tests
	return tableName
}

// CleanupTestTable deletes the test table
func CleanupTestTable(t *testing.T, client *dynamodb.Client, tableName string) {
	t.Helper()

	_, err := client.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		t.Logf("error cleaning up test table: %v", err)
	}
}
