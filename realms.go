package auctionhouse

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func (r *Realm) fillAuctionMap(db *sql.DB) map[int]int {
	storage := make(map[int]int, 0)
	rows, err := db.Query(r.queryString)
	check(err)
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Fatal(err)
		}
		storage[id] = 0

	}
	fmt.Println(len(storage))
	return storage
}

// NewRealm requires the realm name, realm slug and the realm id
func NewRealm(name, slug string, id int, api map[string]string, token string) Realm {
	r := Realm{Name: name, Slug: slug, ID: id}
	r.Region = "us"
	r.Locale = "en_US"
	r.lastChecked = time.Time{}
	r.AuctionURL = r.BuildAuctionURL(api, token)
	r.insertString = fmt.Sprintf(`INSERT INTO "%s"(id, item, orealm, bid, buyout, quantity, timeleft, created) VALUES($1, $2, $3, $4, $5, $6, $7, NOW());`, r.Slug)
	r.queryString = fmt.Sprintf(`select id from "%s";`, r.Slug)
	return r
}

func BuildDBMap() map[string]string {
	strings := make(map[string]interface{}, 0)

	file, err := os.Open("../auctionjson/api.json")
	if err != nil {
		fmt.Println("BuildDBMap() generated an error using os.Open()")
		panic(err)
	}
	body, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("BuildDBMap() generated an error using ioutil.ReadAll()")
		panic(err)
	}
	err = json.Unmarshal(body, &strings)
	if err != nil {
		fmt.Println("BuildDBMap() generated an error using json.Unmarshal()")
		panic(err)
	}
	out := make(map[string]string, 0)
	for k, v := range strings {
		t, ok := v.(string)
		if ok {
			out[k] = t
		} else {
			fmt.Println("BuildDBMap() tried to insert something other than a string in Daemon.API")
		}
	}
	return out
}
func GetFromDBMap(dbAPI map[string]string, strings ...string) (string, bool) {
	var out string
	for _, v := range strings {
		word, ok := dbAPI[v]
		if ok {
			out = out + word
		} else {
			fmt.Println("Error encountered in Daemon.GetAPIStrings()")
			return "", false
		}
	}
	return out, true
}
func (r *Realm) BuildAuctionURL(dbAPI map[string]string, token string) string {
	url, check := GetFromDBMap(dbAPI, "api", "auctionrequest")
	if !check {
		fmt.Println("Error in BuildAuctionURLs()")
	}
	url = strings.Replace(url, regionString, r.Region, 1)
	url = strings.Replace(url, localeString, r.Locale, 1)
	url = strings.Replace(url, tokenString, token, 1)
	url = strings.Replace(url, "{slug}", r.Slug, 1)

	return url
}
