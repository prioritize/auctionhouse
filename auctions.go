package auctionhouse

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// Realm.GetAuctionData
func (r *Realm) GetAuctionData() {

}

func (r *Realm) BuildAuctionURL() {

}
func ServerHandler(r Realms, token string, dbInfo DBInfo, IM *ItemManager) []AuctionHandler {
	auctions := make([]AuctionHandler, 0)
	for _, v := range r.Realms {
		server := NewAuctionHandler(token, v, dbInfo, IM)
		auctions = append(auctions, server)
	}
	return auctions
}
func (a *AuctionHandler) worker() {
	for {
		auction := <-a.Auctions
		a.IM.CheckItem(auction.Item, a.Token)
	}
}
func (a *AuctionHandler) RequestAuctionData() {
	client := http.Client{}
	request, err := http.NewRequest(http.MethodGet, a.URL, nil)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- NewRequest()-1 -- " + a.Realm.Slug)
		return
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- client.Do()-1" + a.Realm.Slug)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- ReadAll() - 1" + a.Realm.Slug)
		return
	}
	files := Files{}
	err = json.Unmarshal(body, &files)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- Unmarshal-1" + a.Realm.Slug)
		return
	}
	request, err = http.NewRequest(http.MethodGet, files.Files[0].URL, nil)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- NewRequest - 2 " + a.Realm.Slug)
		return
	}
	res, err = client.Do(request)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- client.Do() - 2" + a.Realm.Slug)
		return
	}
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- ReadAll() - 2" + a.Realm.Slug)
		log.Fatal(err)
		return
	}
	auctions := Auctions{}
	err = json.Unmarshal(body, &auctions)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- Unmarshal - 2" + a.Realm.Slug)
		return
	}
	for _, v := range auctions.Auctions {
		a.Auctions <- v
	}
}
func NewAuctionHandler(token string, realm Realm, db DBInfo, IM *ItemManager) AuctionHandler {
	a := AuctionHandler{}
	a.Realm = realm
	a.LastChecked = time.Now()
	a.Auctions = make(chan Auction, 5000)
	a.Token = token
	a.URL = realm.URL
	database, ok := OpenDB(db)
	if !ok {
		fmt.Println("Error encountered in NewAuctionHandler()")
		return AuctionHandler{}
	}
	a.db = database
	a.DBInfo = db
	statement := fmt.Sprintf(`INSERT INTO "%s"(id, item, orealm, bid, buyout, quantity, timeleft, created) VALUES($1, $2, $3, $4, $5, $6, $7, NOW());`, a.Realm.Slug)
	insert, err := a.db.Prepare(statement)
	check(err)
	a.Insert = insert
	// !! Open the DB here
	a.IM = IM
	return a
}
func GetDBInfo() (DBInfo, bool) {
	file, err := os.Open("../auctionjson/database.json")
	if err != nil {
		fmt.Println("Error encountered in GetDBInfo()")
		return DBInfo{}, false
	}
	body, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Error encountered in GetDBInfo()")
		return DBInfo{}, false
	}
	db := DBInfo{}
	err = json.Unmarshal(body, &db)
	if err != nil {
		fmt.Println("Error encountered in GetDBInfo()")
		return DBInfo{}, false
	}
	return db, true
}
func OpenDB(db DBInfo) (*sql.DB, bool) {
	psqlConnInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, db.DBname)
	database, err := sql.Open("postgres", psqlConnInfo)
	if err != nil {
		fmt.Println("Error encountered in OpenDB()")
		return &sql.DB{}, false
	}
	return database, true
}

func (a *AuctionHandler) AutomateAuctionCollection() {
	// This function should likely be the only function actually called outside of this file

	// Get the auction URL and check if the lastModified time is different than the currently held value

	// If values are different,

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
	_, err := a.Insert.Exec(auction.AuctionID,
		auction.Item,
		auction.ORealm,
		auction.Bid,
		auction.Buyout,
		auction.Quantity,
		auction.TimeLeft,
	)
	if err != nil {
		fmt.Println("Error from db.Exec() in ParseAuction")
		fmt.Println(err.Error())
	}
}
