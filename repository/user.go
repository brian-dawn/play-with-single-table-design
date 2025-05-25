package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"LearnSingleTableDesign/models"
)

// UserRepository handles User entity operations
type UserRepository struct {
	store *Store
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(client *dynamodb.Client, tableName string) *UserRepository {
	return &UserRepository{
		store: NewStore(client, tableName),
	}
}

// Put stores a user in DynamoDB
func (r *UserRepository) Put(ctx context.Context, user models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	item := GenericItem[models.User]{
		PK:         NewUserPK(user.Email),
		SK:         NewUserSK(user.Email),
		EntityType: EntityUser,
		Data:       user,
	}
	return PutItem(ctx, r.store, item)
}

// Get retrieves a user from DynamoDB
func (r *UserRepository) Get(ctx context.Context, email string) (*models.User, error) {
	var item GenericItem[models.User]
	err := GetItem(ctx, r.store, NewUserPK(email), NewUserSK(email), &item)
	if err != nil {
		return nil, err
	}
	return &item.Data, nil
}
