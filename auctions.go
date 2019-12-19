package auctionhouse

import (
	"auctionauth"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Realm.GetAuctionData
func (r *Realm) GetAuctionData() {

}

func (r *Realm) BuildAuctionURL() {

}
func (a *AuctionHandler) RequestAuctionData() {
	client := http.Client{Timeout: time.Second * 10}
	request, err := http.NewRequest(http.MethodGet, a.URL, nil)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- " + a.Realm.Slug)
		return
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- " + a.Realm.Slug)
	}
	body, err := ioutil.ReadAll(res.Body)
	files := Files{}
	err = json.Unmarshal(body, &files)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- " + a.Realm.Slug)
	}
	request, err = http.NewRequest(http.MethodGet, files.Files[0].URL, nil)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- " + a.Realm.Slug)
		return
	}
	res, err = client.Do(request)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- " + a.Realm.Slug)
	}
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- " + a.Realm.Slug)
	}
	auctions := Auctions{}
	err = json.Unmarshal(body, &auctions)

	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- " + a.Realm.Slug)
	}
	for _, v := range auctions.Auctions {
		a.Auctions <- v
	}
}
func NewAuctionHandler(token auctionauth.Token, realm Realm, insert string) AuctionHandler {
	a := AuctionHandler{}
	a.Realm = realm
	a.LastChecked = time.Now()
	a.Auctions = make(chan Auction, 5000)
	a.Token = token
	a.URL = realm.URL
	a.Insert = "INSERT INTO %s(id, item, orealm, bid, buyout, quantity, timeleft, created) VALUES(%d, %d, '%s', %d, %d, %d, '%s', NOW());"
	// !! Open the DB here
	return a
}

func (a *AuctionHandler) AutomateAuctionCollection() {

}
func (a *AuctionHandler) SendAuctionToDB() {
	for {
		auction, ok := <-a.Auctions
		if ok {
			fmt.Println("Something happened")
			a.ParseAuction(auction)
		}
	}
}

func (a *AuctionHandler) ParseAuction(auction Auction) {
	fmt.Println(auction)
	fmt.Println(a.Insert)
	fmt.Println(a.Realm.Slug)
	fmt.Println(auction.Item)
	fmt.Println(auction.ORealm)
	fmt.Println(auction.Bid)
	fmt.Println(auction.Buyout)
	fmt.Println(auction.Quantity)
	fmt.Println(auction.TimeLeft)
	res, err := a.db.Exec(a.Insert, a.Realm.Slug,
		auction.AuctionID,
		auction.Item,
		auction.ORealm,
		auction.Bid,
		auction.Buyout,
		auction.Quantity,
		auction.TimeLeft,
	)
	fmt.Println(res)
	if err != nil {
		fmt.Println(res)
	}
}
