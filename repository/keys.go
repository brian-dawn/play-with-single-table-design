package repository

import "fmt"

type KeyFactory struct{}

var Key = KeyFactory{}

// Key constructors
func (KeyFactory) NewUserPK(email string) PrimaryKey {
	return PrimaryKey(fmt.Sprintf("USER#%s", email))
}

func (KeyFactory) NewUserSK(email string) SortKey {
	return SortKey(fmt.Sprintf("PROFILE#%s", email))
}

func (KeyFactory) NewOrderSK(orderID string) SortKey {
	return SortKey(fmt.Sprintf("ORDER#%s", orderID))
}

func (KeyFactory) NewProductPK() PrimaryKey {
	return PrimaryKey("PRODUCT#ALL")
}

func (KeyFactory) NewProductSK(productID string) SortKey {
	return SortKey(fmt.Sprintf("PRODUCT#%s", productID))
}
