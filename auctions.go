package auctionhouse

import (
	"auctionauth"
	"time"
)

// Realm.GetAuctionData
func (r *Realm) GetAuctionData() {

}

func (r *Realm) BuildAuctionURL() {

}
func RequestAuctionData() {}
func NewAuctionHandler(token auctionauth.Token, realm Realm, insert string) AuctionHandler {
	a := AuctionHandler{}
	a.Realm = realm
	a.LastChecked = time.Now()
	a.Insert = insert
	a.Auctions = make(chan Auction, 5000)
	a.Token = token
	return a
}
