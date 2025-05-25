package repository

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"LearnSingleTableDesign/models"
	"LearnSingleTableDesign/testutil"
)

// testSetup creates test resources and returns cleanup function
func testSetup(t *testing.T) (*dynamodb.Client, string, *UserRepository, *OrderRepository, *ProductRepository, func()) {
	t.Helper()
	client := testutil.CreateTestClient(t)
	tableName := testutil.SetupTestTable(t, client)

	userRepo := NewUserRepository(client, tableName)
	orderRepo := NewOrderRepository(client, tableName)
	productRepo := NewProductRepository(client, tableName)

	cleanup := func() {
		testutil.CleanupTestTable(t, client, tableName)
	}

	return client, tableName, userRepo, orderRepo, productRepo, cleanup
}

// createTestData creates test data for use in tests
func createTestData() (models.User, []models.Order, []models.Product) {
	testProducts := []models.Product{
		{
			ProductID: "PROD1",
			Name:      "Product 1",
			Category:  "Electronics",
			Price:     100.00,
			Stock:     100,
			CreatedAt: time.Now(),
		},
		{
			ProductID: "PROD2",
			Name:      "Product 2",
			Category:  "Electronics",
			Price:     200.00,
			Stock:     100,
			CreatedAt: time.Now(),
		},
	}

	testUser := models.User{
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	testOrders := []models.Order{
		{
			OrderID:   "ORD1",
			UserEmail: testUser.Email,
			Status:    "PENDING",
			Total:     99.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD1"},
		},
		{
			OrderID:   "ORD2",
			UserEmail: testUser.Email,
			Status:    "COMPLETED",
			Total:     199.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD2", "PROD3"},
		},
		{
			OrderID:   "ORD3",
			UserEmail: testUser.Email,
			Status:    "PENDING",
			Total:     299.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD4"},
		},
	}

	return testUser, testOrders, testProducts
}

func TestUserRepository_Put(t *testing.T) {
	_, _, userRepo, _, _, cleanup := testSetup(t)
	defer cleanup()

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
		t.Run(tt.name, func(t *testing.T) {
			err := userRepo.Put(context.Background(), tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the user was stored correctly
				got, err := userRepo.Get(context.Background(), tt.user.Email)
				if err != nil {
					t.Errorf("Get() error = %v", err)
					return
				}
				if got.Email != tt.user.Email {
					t.Errorf("Email = %v, want %v", got.Email, tt.user.Email)
				}
				if got.Name != tt.user.Name {
					t.Errorf("Name = %v, want %v", got.Name, tt.user.Name)
				}
				if got.CreatedAt.Sub(tt.user.CreatedAt) > time.Second {
					t.Errorf("CreatedAt = %v, want %v (within 1s)", got.CreatedAt, tt.user.CreatedAt)
				}
			}
		})
	}
}

func TestUserRepository_Get(t *testing.T) {
	_, _, userRepo, _, _, cleanup := testSetup(t)
	defer cleanup()

	testUser, _, _ := createTestData()

	// Store test user
	err := userRepo.Put(context.Background(), testUser)
	if err != nil {
		t.Fatalf("Failed to put test user: %v", err)
	}

	tests := []struct {
		name    string
		email   string
		want    *models.User
		wantErr bool
	}{
		{
			name:  "existing user",
			email: testUser.Email,
			want: &models.User{
				Email:     testUser.Email,
				Name:      testUser.Name,
				CreatedAt: testUser.CreatedAt,
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
		t.Run(tt.name, func(t *testing.T) {
			got, err := userRepo.Get(context.Background(), tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Email != tt.want.Email {
					t.Errorf("Email = %v, want %v", got.Email, tt.want.Email)
				}
				if got.Name != tt.want.Name {
					t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
				}
				if got.CreatedAt.Sub(tt.want.CreatedAt) > time.Second {
					t.Errorf("CreatedAt = %v, want %v (within 1s)", got.CreatedAt, tt.want.CreatedAt)
				}
			}
		})
	}
}

func TestProductRepository_Put(t *testing.T) {
	_, _, _, _, productRepo, cleanup := testSetup(t)
	defer cleanup()

	_, _, testProducts := createTestData()

	// For each product perform put
	for _, product := range testProducts {
		err := productRepo.Put(context.Background(), product)
		if err != nil {
			t.Errorf("Put() error = %v", err)
			continue
		}

		// Make sure we can get as well
		got, err := productRepo.Get(context.Background(), product.ProductID)
		if err != nil {
			t.Errorf("Get() error = %v", err)
			continue
		}

		if got.ProductID != product.ProductID {
			t.Errorf("ProductID = %v, want %v", got.ProductID, product.ProductID)
		}
		if got.Name != product.Name {
			t.Errorf("Name = %v, want %v", got.Name, product.Name)
		}
		if got.Category != product.Category {
			t.Errorf("Category = %v, want %v", got.Category, product.Category)
		}
		if got.Price != product.Price {
			t.Errorf("Price = %v, want %v", got.Price, product.Price)
		}
		if got.Stock != product.Stock {
			t.Errorf("Stock = %v, want %v", got.Stock, product.Stock)
		}
		if got.CreatedAt.Sub(product.CreatedAt) > time.Second {
			t.Errorf("CreatedAt = %v, want %v (within 1s)", got.CreatedAt, product.CreatedAt)
		}
	}
}

func TestOrderRepository_Put(t *testing.T) {
	_, _, _, orderRepo, _, cleanup := testSetup(t)
	defer cleanup()

	_, testOrders, _ := createTestData()

	tests := []struct {
		name    string
		order   models.Order
		wantErr bool
	}{
		{
			name:    "valid order",
			order:   testOrders[0],
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
		t.Run(tt.name, func(t *testing.T) {
			err := orderRepo.Put(context.Background(), tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderRepository_GetUserOrders(t *testing.T) {
	_, _, _, orderRepo, _, cleanup := testSetup(t)
	defer cleanup()

	testUser, testOrders, _ := createTestData()

	// Store test orders
	for _, order := range testOrders {
		err := orderRepo.Put(context.Background(), order)
		if err != nil {
			t.Fatalf("Failed to put test order: %v", err)
		}
	}

	tests := []struct {
		name          string
		userEmail     string
		opts          *QueryOptions
		wantCount     int
		wantNextToken bool
		wantErr       bool
	}{
		{
			name:      "get all orders",
			userEmail: testUser.Email,
			opts:      nil,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "get orders with pagination",
			userEmail: testUser.Email,
			opts: &QueryOptions{
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
		t.Run(tt.name, func(t *testing.T) {
			result, err := orderRepo.GetUserOrders(context.Background(), tt.userEmail, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result.Orders) != tt.wantCount {
				t.Errorf("Got %d orders, want %d", len(result.Orders), tt.wantCount)
			}

			if (result.NextPageToken != nil) != tt.wantNextToken {
				t.Errorf("NextPageToken = %v, want %v", result.NextPageToken != nil, tt.wantNextToken)
			}

			// Verify order fields
			if tt.wantCount > 0 {
				for _, order := range result.Orders {
					if order.UserEmail != tt.userEmail {
						t.Errorf("UserEmail = %v, want %v", order.UserEmail, tt.userEmail)
					}
					if order.OrderID == "" {
						t.Error("OrderID is empty")
					}
					if order.Status == "" {
						t.Error("Status is empty")
					}
					if order.Total <= 0 {
						t.Error("Total is not positive")
					}
					if len(order.Products) == 0 {
						t.Error("Products is empty")
					}
				}
			}
		})
	}
}
