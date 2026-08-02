package shop

import (
	"strconv"
	"strings"
)

func validateShopItemName(name string) bool {
	return strings.TrimSpace(name) != ""
}

func validateShopItemPrice(price string) bool {
	p, err := strconv.ParseInt(price, 10, 32)
	if err != nil {
		return false
	}
	return p > 0
}
