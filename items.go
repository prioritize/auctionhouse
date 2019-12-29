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
	"strings"
	"time"
)

func NewItemManager() ItemManager {
	i := ItemManager{}
	i.api = loadAPI()
	i.toAdd = make(chan Item, 500)
	i.Items = make(map[int]int, 0)
	i.fillDBInfo()
	i.openDB()
	statement, err := i.db.Prepare("select * from items where item=$1")
	check(err)
	i.QueryStatement = statement
	item := Item{}
	item.ID = 153604
	i.checkDBForItem(item)
	// go i.insertProcess()
	// go i.queryProcess()
	return i
}

func (i *ItemManager) openDB() {
	psqlConnInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		i.dbInfo.Host, i.dbInfo.Port, i.dbInfo.User, i.dbInfo.Password, i.dbInfo.DBname)
	database, err := sql.Open("postgres", psqlConnInfo)
	check(err)
	i.db = database
}

// !! ProcessItem should
// -- Check the database for the item
// -- If the item exists, return
// -- if the item does not exist, query the blizzard api
// -- input the data into the database
// -- return
func (i *ItemManager) processItem(item Item) {

}

// QueryItemInformation requests Item information from the Blizzard API and return an Item to be placed into the database
func (i *ItemManager) QueryItemInformation(item int, token Token) (Item, bool) {
	url := i.GetItemURL(item, token)
	client := http.Client{Timeout: timeout * time.Second}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	check(err)
	res, err := client.Do(request)
	check(err)
	if res.StatusCode != 200 {
		return Item{}, false
	}
	body, err := ioutil.ReadAll(res.Body)
	check(err)
	newItem := Item{}
	err = json.Unmarshal(body, &newItem)
	check(err)
	return newItem, true
}
func (i *ItemManager) queryProcess() {
	for {
		// value := <-i.toQueryAPI
	}

}

func (i *ItemManager) insertProcess() {

}

// GetAPIStrings accepts strings that should be found in the AH API and returns them concatenated together
func (i *ItemManager) GetAPIStrings(components ...string) (string, bool) {
	var out string
	for _, v := range components {
		word, ok := i.api[v]
		if ok {
			out = out + word
		} else {
			fmt.Println("Error encountered in Daemon.GetAPIStrings()")
			return "", false
		}
	}
	return out, true
}
func loadAPI() map[string]string {
	strings := make(map[string]interface{}, 0)

	file, err := os.Open("../auctionjson/api.json")
	if err != nil {
		fmt.Println("LoadMapWithAPI() generated an error using os.Open()")
		panic(err)
	}
	body, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("LoadMapWithAPI() generated an error using ioutil.ReadAll()")
		panic(err)
	}
	err = json.Unmarshal(body, &strings)
	if err != nil {
		fmt.Println("LoadMapWithAPI() generated an error using json.Unmarshal()")
		panic(err)
	}
	out := make(map[string]string, 0)
	for k, v := range strings {
		t, ok := v.(string)
		if ok {
			out[k] = t
		} else {
			fmt.Println("LoadMapWithAPI() tried to insert something other than a string in Daemon.API")
		}
	}
	return out
}

// GetItemURL returns the Blizzard API URL to request the item  from the Blizzard API
func (i *ItemManager) GetItemURL(item int, token Token) string {
	url, check := i.GetAPIStrings("itemrequest")
	if !check {
		fmt.Println("URL wasn't found")
	}
	itemString := strconv.Itoa(item)
	url = strings.Replace(url, regionString, "us", 1)
	url = strings.Replace(url, localeString, "en_US", 1)
	url = strings.Replace(url, tokenString, token.Token(), 1)
	url = strings.Replace(url, "{item}", itemString, 1)
	return url
}
func (i *ItemManager) checkDBForItem(item Item) {
	rows, err := i.QueryStatement.Query(item.ID)
	check(err)
	defer rows.Close()
	for rows.Next() {
		var name string
		// !! Need to make this scan an Item, not a string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		fmt.Println("This is the name!")
		fmt.Println(name)
	}
}

func (i *ItemManager) fillDBInfo() {
	file, err := os.Open("../auctionjson/database.json")
	check(err)
	body, err := ioutil.ReadAll(file)
	check(err)
	dbInfo := DBInfo{}
	err = json.Unmarshal(body, &dbInfo)
	check(err)
	i.dbInfo = dbInfo
}
