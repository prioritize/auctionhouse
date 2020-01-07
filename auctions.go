package auctionhouse

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	// this is required to use the psql driver
	_ "github.com/lib/pq"
)

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
func NewAuctionHandler(token string, realm Realm, db *sql.DB, IM *ItemManager) AuctionHandler {
	a := AuctionHandler{}
	a.Realm = realm
	a.LastChecked = time.Time{}
	a.Auctions = make(chan Auction, 50000)
	a.Token = token
	a.URL = realm.URL
	dbInfo, ok := GetDBInfo()
	if !ok {
		fmt.Println("Couldn't get DBInfo in NewAuctionHandler()")
	}
	a.DBInfo = dbInfo
	// database, ok := OpenDB(db)
	a.db = db
	statement := fmt.Sprintf(`INSERT INTO "%s"(id, item, orealm, bid, buyout, quantity, timeleft, created) VALUES($1, $2, $3, $4, $5, $6, $7, NOW());`, a.Realm.Slug)
	insert, err := a.db.Prepare(statement)
	check(err)
	a.Insert = insert
	// !! Open the DB here
	a.IM = IM
	go a.SendAuctionToDB()
	go a.SendAuctionToDB()
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

func (a *AuctionHandler) SendAuctionToDB() {
	for {
		auction, ok := <-a.Auctions
		if ok {
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
	if err == nil {
		// this is likely going to generate a race condition, but that isn't a huge issue
		a.count++
	} else {
		fmt.Println(err.Error())
	}
	if a.count > 0 && a.count%100 == 0 {
		fmt.Println("Keep It 100" + strconv.Itoa(a.count))
		// fmt.Println(info.LastInsertId())
	}
}
