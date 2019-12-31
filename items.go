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
	i.toAdd = make(chan Item, 50000)
	i.toQuery = make(chan ItemQuery, 50000)
	i.Items = make(map[int]int, 0)
	i.fillDBInfo()
	i.openDB()
	query, err := i.db.Prepare("select * from items where item=$1")
	check(err)
	i.QueryStatement = query
	statement, err := i.db.Prepare("INSERT INTO items(item, href, name) values($1, $2, $3)")
	check(err)
	i.InsertStatement = statement
	go i.insertWorker()
	go i.queryWorker()
	return i
}

func (i *ItemManager) openDB() {
	psqlConnInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		i.DBInfo.Host, i.DBInfo.Port, i.DBInfo.User, i.DBInfo.Password, i.DBInfo.DBname)
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
func (i *ItemManager) QueryItemInformation(item int, token string) (Item, bool) {
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
func (i *ItemManager) checkDBForItem(item int) bool {
	rows, err := i.QueryStatement.Query(item)
	check(err)
	defer rows.Close()
	var found bool
	for rows.Next() {
		retrievedItem := Item{}
		retrievedItem.Icon.Asset = make([]Asset, 1)
		found = true
		// !! Need to make this scan an Item, not a string
		if err := rows.Scan(&retrievedItem.ID, &retrievedItem.Icon.Asset[0].HREF, &retrievedItem.Name); err != nil {
			log.Fatal(err)
		}
	}
	return found
}
func (i *ItemManager) queryWorker() {
	rate := time.Millisecond * 10
	throttle := time.Tick(rate)
	for item := range i.toQuery {
		value := i.checkDBForItem(item.Item)
		fmt.Println(value)
		if !value {
			<-throttle
			newItem := i.NewItem(item.Item, item.Token)
			i.toAdd <- newItem
		}
	}
}
func (i *ItemManager) insertItemInDB(item Item) {
	i.InsertStatement.Exec(item.ID, item.Icon.Asset[0].HREF, item.Name)
}

func (i *ItemManager) insertWorker() {
	for item := range i.toAdd {
		// newItem := i.NewItem(item.)
		i.insertItemInDB(item)
	}
}
func (i *ItemManager) CheckItem(item int, token string) {
	i.toQuery <- ItemQuery{Item: item, Token: token}
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
func (i *ItemManager) GetItemURL(item int, token string) string {
	url, check := i.GetAPIStrings("itemrequest")
	if !check {
		fmt.Println("URL wasn't found")
	}
	itemString := strconv.Itoa(item)
	url = strings.Replace(url, regionString, "us", 1)
	url = strings.Replace(url, localeString, "en_US", 1)
	url = strings.Replace(url, tokenString, token, 1)
	url = strings.Replace(url, "{item}", itemString, 1)
	return url
}

func (i *ItemManager) fillDBInfo() {
	file, err := os.Open("../auctionjson/database.json")
	check(err)
	body, err := ioutil.ReadAll(file)
	check(err)
	dbInfo := DBInfo{}
	err = json.Unmarshal(body, &dbInfo)
	check(err)
	i.DBInfo = dbInfo
}
func (i *ItemManager) NewItem(item int, token string) Item {
	url := i.GetItemURL(item, token)
	newItem := i.GetItem(url, token)
	return newItem
}
func (i *ItemManager) GetItem(url string, token string) Item {
	client := http.Client{Timeout: timeout * time.Second}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	check(err)
	response, err := client.Do(request)
	check(err)
	body, err := ioutil.ReadAll(response.Body)
	check(err)
	item := Item{}
	err = json.Unmarshal(body, &item)
	check(err)
	mediaURL := item.Media.Key.HREF + "&access_token=" + token
	mediaRequest, err := http.NewRequest(http.MethodGet, mediaURL, nil)
	check(err)
	response, err = client.Do(mediaRequest)
	check(err)
	body, err = ioutil.ReadAll(response.Body)
	check(err)
	icon := Icon{}
	err = json.Unmarshal(body, &icon)
	fmt.Println(item)
	item.Icon = icon
	return item
}

// Close closes the two channels in the ItemManager which terminates the worker processes
func (i *ItemManager) Close() {
	close(i.toAdd)
	close(i.toQuery)
}
