package repository

import "fmt"

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

func NewCategoryPK() PrimaryKey {
	return PrimaryKey("PRODUCT#ALL")
}

func NewProductSK(productID string) SortKey {
	return SortKey(fmt.Sprintf("PRODUCT#%s", productID))
}
