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

	// Test putting a valid user
	user := models.User{
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	err := userRepo.Put(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to put valid user: %v", err)
	}

	// Verify the user was stored correctly
	got, err := userRepo.Get(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("Failed to get user after put: %v", err)
	}

	if got.Email != user.Email {
		t.Errorf("Email = %v, want %v", got.Email, user.Email)
	}
	if got.Name != user.Name {
		t.Errorf("Name = %v, want %v", got.Name, user.Name)
	}
	if got.CreatedAt.Sub(user.CreatedAt) > time.Second {
		t.Errorf("CreatedAt = %v, want %v (within 1s)", got.CreatedAt, user.CreatedAt)
	}

	// Test putting an invalid user (missing email)
	invalidUser := models.User{
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	err = userRepo.Put(context.Background(), invalidUser)
	if err == nil {
		t.Error("Expected error when putting user with missing email, got nil")
	}

	// Test putting an invalid user (missing name)
	invalidUser = models.User{
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	err = userRepo.Put(context.Background(), invalidUser)
	if err == nil {
		t.Error("Expected error when putting user with missing name, got nil")
	}
}

func TestUserRepository_Get(t *testing.T) {
	_, _, userRepo, _, _, cleanup := testSetup(t)
	defer cleanup()

	// Create and store a test user
	user := models.User{
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	err := userRepo.Put(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to put test user: %v", err)
	}

	// Test getting an existing user
	got, err := userRepo.Get(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("Failed to get existing user: %v", err)
	}

	if got.Email != user.Email {
		t.Errorf("Email = %v, want %v", got.Email, user.Email)
	}
	if got.Name != user.Name {
		t.Errorf("Name = %v, want %v", got.Name, user.Name)
	}
	if got.CreatedAt.Sub(user.CreatedAt) > time.Second {
		t.Errorf("CreatedAt = %v, want %v (within 1s)", got.CreatedAt, user.CreatedAt)
	}

	// Test getting a non-existent user
	_, err = userRepo.Get(context.Background(), "nonexistent@example.com")
	if err == nil {
		t.Error("Expected error when getting non-existent user, got nil")
	}
}

func TestProductRepository_Put(t *testing.T) {
	_, _, _, _, productRepo, cleanup := testSetup(t)
	defer cleanup()

	// Create and store a test product
	product := models.Product{
		ProductID: "PROD1",
		Name:      "Test Product",
		Category:  "Electronics",
		Price:     100.00,
		Stock:     100,
		CreatedAt: time.Now(),
	}

	err := productRepo.Put(context.Background(), product)
	if err != nil {
		t.Fatalf("Failed to put product: %v", err)
	}

	// Verify the product was stored correctly
	got, err := productRepo.Get(context.Background(), product.ProductID)
	if err != nil {
		t.Fatalf("Failed to get product after put: %v", err)
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

func TestOrderRepository_Put(t *testing.T) {
	_, _, _, orderRepo, _, cleanup := testSetup(t)
	defer cleanup()

	// Test putting a valid order
	order := models.Order{
		OrderID:   "ORD1",
		UserEmail: "test@example.com",
		Status:    "PENDING",
		Total:     99.99,
		CreatedAt: time.Now(),
		Products:  []string{"PROD1"},
	}

	err := orderRepo.Put(context.Background(), order)
	if err != nil {
		t.Fatalf("Failed to put valid order: %v", err)
	}

	// Test putting an invalid order (missing order ID)
	invalidOrder := models.Order{
		UserEmail: "test@example.com",
		Status:    "PENDING",
		Total:     99.99,
		CreatedAt: time.Now(),
		Products:  []string{"PROD1"},
	}

	err = orderRepo.Put(context.Background(), invalidOrder)
	if err == nil {
		t.Error("Expected error when putting order with missing order ID, got nil")
	}

	// Test putting an invalid order (missing user email)
	invalidOrder = models.Order{
		OrderID:   "ORD123",
		Status:    "PENDING",
		Total:     99.99,
		CreatedAt: time.Now(),
		Products:  []string{"PROD1"},
	}

	err = orderRepo.Put(context.Background(), invalidOrder)
	if err == nil {
		t.Error("Expected error when putting order with missing user email, got nil")
	}
}

func TestOrderRepository_GetUserOrders(t *testing.T) {
	_, _, _, orderRepo, _, cleanup := testSetup(t)
	defer cleanup()

	userEmail := "test@example.com"

	// Create and store some test orders
	orders := []models.Order{
		{
			OrderID:   "ORD1",
			UserEmail: userEmail,
			Status:    "PENDING",
			Total:     99.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD1"},
		},
		{
			OrderID:   "ORD2",
			UserEmail: userEmail,
			Status:    "COMPLETED",
			Total:     199.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD2", "PROD3"},
		},
		{
			OrderID:   "ORD3",
			UserEmail: userEmail,
			Status:    "PENDING",
			Total:     299.99,
			CreatedAt: time.Now(),
			Products:  []string{"PROD4"},
		},
	}

	for _, order := range orders {
		err := orderRepo.Put(context.Background(), order)
		if err != nil {
			t.Fatalf("Failed to put test order: %v", err)
		}
	}

	// Test getting all orders
	result, err := orderRepo.GetUserOrders(context.Background(), userEmail, nil)
	if err != nil {
		t.Fatalf("Failed to get user orders: %v", err)
	}

	if len(result.Orders) != len(orders) {
		t.Errorf("Got %d orders, want %d", len(result.Orders), len(orders))
	}

	// Test pagination
	result, err = orderRepo.GetUserOrders(context.Background(), userEmail, &QueryOptions{Limit: 2})
	if err != nil {
		t.Fatalf("Failed to get paginated user orders: %v", err)
	}

	if len(result.Orders) != 2 {
		t.Errorf("Got %d orders with pagination, want 2", len(result.Orders))
	}

	if result.NextPageToken == nil {
		t.Error("Expected next page token for paginated results, got nil")
	}

	// Test getting orders for non-existent user
	result, err = orderRepo.GetUserOrders(context.Background(), "nonexistent@example.com", nil)
	if err != nil {
		t.Fatalf("Failed to get orders for non-existent user: %v", err)
	}

	if len(result.Orders) != 0 {
		t.Errorf("Got %d orders for non-existent user, want 0", len(result.Orders))
	}
}
