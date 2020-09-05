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

func (d *Daemon) AuctionWorker() {
	tick := time.Tick(time.Second * 3)
	for {
		<-tick
		fmt.Println("len(realms): " + strconv.Itoa(len(d.realms)))
		fmt.Println("len(dbPool): " + strconv.Itoa(len(d.dbPool)))
		realm := <-d.realms
		db := <-d.dbPool
		fmt.Println(realm.Slug + " last modified before: " + strconv.Itoa(realm.lastModified))
		d.RequestAuctionData(&realm, db)
		fmt.Println(realm.Slug + " last modified after: " + strconv.Itoa(realm.lastModified))
		fmt.Println()
		d.dbPool <- db
		d.realms <- realm
	}
}
func (d *Daemon) RequestAuctionData(r *Realm, db *sql.DB) {
	client := <-d.httpPool
	request, err := http.NewRequest(http.MethodGet, r.AuctionURL, nil)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- NewRequest()-1 -- " + r.Slug)
		return
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- client.Do()-1" + r.Slug)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- ReadAll() - 1" + r.Slug)
		return
	}
	files := Files{}
	err = json.Unmarshal(body, &files)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- Unmarshal-1" + r.Slug)
		log.Fatal(err)
		return
	}
	if files.Files[0].Modified == r.lastModified {
		fmt.Println(r.Slug + " auctions not modified, skipping")
		return
	}
	r.lastModified = files.Files[0].Modified
	request, err = http.NewRequest(http.MethodGet, files.Files[0].URL, nil)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- NewRequest - 2 " + r.Slug)
		return
	}
	res, err = client.Do(request)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- client.Do() - 2" + r.Slug)
		return
	}
	fmt.Println("res.StatusCode: " + strconv.Itoa(res.StatusCode))
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- ReadAll() - 2" + r.Slug)
		log.Fatal(err)
		return
	}
	auctions := Auctions{}
	err = json.Unmarshal(body, &auctions)
	if err != nil {
		fmt.Println("Error in RequestAuctionData() -- Unmarshal - 2" + r.Slug)
		d.httpPool <- client
		return
	}
	storage := r.fillAuctionMap(db)
	for _, v := range auctions.Auctions {
		_, ok := storage[v.AuctionID]
		if !ok {
			d.Auctions <- v
		}
	}
	fmt.Println("-----------------------------------------------------")
	fmt.Println("All auctions from " + r.Slug + " sent to channel")
	fmt.Println("Number of items in auction channel: " + strconv.Itoa(len(d.Auctions)))
	fmt.Println("-----------------------------------------------------")
	fmt.Println()
	d.httpPool <- client
}

// func NewAuctionHandler(token string, realm Realm, db *sql.DB, IM *ItemManager) AuctionHandler {
// 	a := AuctionHandler{}
// 	a.Realm = realm
// 	a.LastChecked = time.Time{}
// 	a.Auctions = make(chan Auction, 50000)
// 	a.Token = token
// 	a.URL = realm.URL
// 	dbInfo, ok := GetDBInfo()
// 	if !ok {
// 		fmt.Println("Couldn't get DBInfo in NewAuctionHandler()")
// 	}
// 	a.DBInfo = dbInfo
// 	// database, ok := OpenDB(db)
// 	a.db = db
// 	statement := fmt.Sprintf(`INSERT INTO "%s"(id, item, orealm, bid, buyout, quantity, timeleft, created) VALUES($1, $2, $3, $4, $5, $6, $7, NOW());`, a.Realm.Slug)
// 	insert, err := a.db.Prepare(statement)
// 	check(err)
// 	statement = fmt.Sprintf(`select id from "%s";`, a.Realm.Slug)
// 	query, err := a.db.Prepare(statement)
// 	check(err)
// 	a.Insert = insert
// 	a.Query = query
// 	// !! Open the DB here
// 	a.IM = IM
// 	return a
// }
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

// func (a *AuctionHandler) SendAuctionToDB(auction Auction, db *sql.DB) {
// 	db.InsertAuction.Exec(auction.AuctionID,
// 		auction.Item,
// 		auction.ORealm,
// 		auction.Bid,
// 		auction.Buyout,
// 		auction.Quantity,
// 		auction.TimeLeft,
// 	)
// }
