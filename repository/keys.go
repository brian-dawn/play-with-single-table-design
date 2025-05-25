package repository

import "fmt"

type KeyFactory struct{}

var Key = KeyFactory{}

func (KeyFactory) UserPK(email string) PrimaryKey {
	return PrimaryKey(fmt.Sprintf("USER#%s", email))
}

func (KeyFactory) UserSK(email string) SortKey {
	return SortKey(fmt.Sprintf("PROFILE#%s", email))
}

func (KeyFactory) OrderSK(orderID string) SortKey {
	return SortKey(fmt.Sprintf("ORDER#%s", orderID))
}

func (KeyFactory) ProductPK() PrimaryKey {
	return "PRODUCT#ALL"
}

func (KeyFactory) ProductSK(productID string) SortKey {
	return SortKey(fmt.Sprintf("PRODUCT#%s", productID))
}
