package tests

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/suite"

	"LearnSingleTableDesign/internal/datastore"
	"LearnSingleTableDesign/internal/models"
	"LearnSingleTableDesign/internal/testutil"
)

// RepositoryTestSuite defines a test suite for repository tests
type RepositoryTestSuite struct {
	suite.Suite
	client     *dynamodb.Client
	tableName  string
	userRepo   *datastore.UserRepository
	orderRepo  *datastore.OrderRepository
	testUser   models.User
	testOrders []models.Order
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) SetupSuite() {
	s.client = testutil.CreateTestClient(s.T())
}

func (s *RepositoryTestSuite) SetupTest() {
	s.tableName = testutil.SetupTestTable(s.T(), s.client)
	s.userRepo = datastore.NewUserRepository(s.client, s.tableName)
	s.orderRepo = datastore.NewOrderRepository(s.client, s.tableName)

	// Create test user
	s.testUser = models.User{
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	// Create test orders
	s.testOrders = []models.Order{
		{
			OrderID:   "ORD1",
			UserEmail: s.testUser.Email,
			Status:    "PENDING",
			Total:     99.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD1"},
		},
		{
			OrderID:   "ORD2",
			UserEmail: s.testUser.Email,
			Status:    "COMPLETED",
			Total:     199.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD2", "PROD3"},
		},
		{
			OrderID:   "ORD3",
			UserEmail: s.testUser.Email,
			Status:    "PENDING",
			Total:     299.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD4"},
		},
	}
}

func (s *RepositoryTestSuite) TearDownTest() {
	testutil.CleanupTestTable(s.T(), s.client, s.tableName)
}

// User Repository Tests

func (s *RepositoryTestSuite) TestUserRepository_Put() {
	tests := []struct {
		name    string
		user    models.User
		wantErr bool
	}{
		{
			name: "valid user",
			user: models.User{
				Email:     "test@example.com",
				Name:      "Test User",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid user - missing email",
			user: models.User{
				Name:      "Test User",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid user - missing name",
			user: models.User{
				Email:     "test@example.com",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.userRepo.Put(context.Background(), tt.user)
			if tt.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)

			// Verify the user was stored correctly
			got, err := s.userRepo.Get(context.Background(), tt.user.Email)
			s.Require().NoError(err)
			s.Equal(tt.user.Email, got.Email)
			s.Equal(tt.user.Name, got.Name)
			s.WithinDuration(tt.user.CreatedAt, got.CreatedAt, time.Second)
		})
	}
}

func (s *RepositoryTestSuite) TestUserRepository_Get() {
	// Store test user
	err := s.userRepo.Put(context.Background(), s.testUser)
	s.Require().NoError(err)

	tests := []struct {
		name    string
		email   string
		want    *models.User
		wantErr bool
	}{
		{
			name:  "existing user",
			email: s.testUser.Email,
			want: &models.User{
				Email:     s.testUser.Email,
				Name:      s.testUser.Name,
				CreatedAt: s.testUser.CreatedAt,
			},
			wantErr: false,
		},
		{
			name:    "non-existent user",
			email:   "nonexistent@example.com",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := s.userRepo.Get(context.Background(), tt.email)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.Require().NoError(err)
			s.Equal(tt.want.Email, got.Email)
			s.Equal(tt.want.Name, got.Name)
			s.WithinDuration(tt.want.CreatedAt, got.CreatedAt, time.Second)
		})
	}
}

// Order Repository Tests

func (s *RepositoryTestSuite) TestOrderRepository_Put() {
	tests := []struct {
		name    string
		order   models.Order
		wantErr bool
	}{
		{
			name:    "valid order",
			order:   s.testOrders[0],
			wantErr: false,
		},
		{
			name: "invalid order - missing order ID",
			order: models.Order{
				UserEmail: "test@example.com",
				Status:    "PENDING",
				Total:     99.99,
				CreatedAt: time.Now(),
				Products:  []string{"PROD1"},
			},
			wantErr: true,
		},
		{
			name: "invalid order - missing user email",
			order: models.Order{
				OrderID:   "ORD123",
				Status:    "PENDING",
				Total:     99.99,
				CreatedAt: time.Now(),
				Products:  []string{"PROD1"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.orderRepo.Put(context.Background(), tt.order)
			if tt.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)
		})
	}
}

func (s *RepositoryTestSuite) TestOrderRepository_GetUserOrders() {
	// Store test orders
	for _, order := range s.testOrders {
		err := s.orderRepo.Put(context.Background(), order)
		s.Require().NoError(err)
	}

	tests := []struct {
		name          string
		userEmail     string
		opts          *datastore.QueryOptions
		wantCount     int
		wantNextToken bool
		wantErr       bool
	}{
		{
			name:      "get all orders",
			userEmail: s.testUser.Email,
			opts:      nil,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "get orders with pagination",
			userEmail: s.testUser.Email,
			opts: &datastore.QueryOptions{
				Limit: 2,
			},
			wantCount:     2,
			wantNextToken: true,
			wantErr:       false,
		},
		{
			name:      "get orders for non-existent user",
			userEmail: "nonexistent@example.com",
			opts:      nil,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := s.orderRepo.GetUserOrders(context.Background(), tt.userEmail, tt.opts)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.Require().NoError(err)
			s.Len(result.Orders, tt.wantCount)
			s.Equal(tt.wantNextToken, result.NextPageToken != nil)

			// Verify order fields
			if tt.wantCount > 0 {
				for _, order := range result.Orders {
					s.Equal(tt.userEmail, order.UserEmail)
					s.NotEmpty(order.OrderID)
					s.NotEmpty(order.Status)
					s.Greater(order.Total, float64(0))
					s.NotEmpty(order.Products)
				}
			}
		})
	}
}
