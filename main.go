package auctionhouse

import (
	"auctionauth"
	"fmt"
)

func main() {
	token, check := auctionauth.NewTokenData()
	if !check {
		fmt.Println(token)
	} else {
		fmt.Println("NewTokenData failed")
	}
}
