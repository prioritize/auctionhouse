package auctionhouse

import (
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

func NewDaemon(region, locale string) (Daemon, bool) {
	d := Daemon{Region: region, Locale: locale}
	d.LoadMapWithAPI()
	t := NewToken()
	d.Token = t
	dbInfo, ok := GetDBInfo()
	if !ok {
		fmt.Println("Failed to create NewDaemon()")
		return Daemon{}, false
	}
	db, ok := OpenDB(dbInfo)
	if !ok {
		fmt.Println("Failed to create NewDaemon()")
		return Daemon{}, false
	}
	realms, ok := d.GetRealms()
	itemManager := NewItemManager()
	AM := make([]AuctionHandler, 0)
	for _, v := range realms.Realms {
		handler := NewAuctionHandler(d.Token.Token(), v, db, &itemManager)
		AM = append(AM, handler)
	}
	d.AuctionManager = AM
	go d.monitor()
	return d, true
}
func (d *Daemon) LoadMapWithAPI() {
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
	d.API = make(map[string]string, 0)
	for k, v := range strings {
		t, ok := v.(string)
		if ok {
			d.API[k] = t
		} else {
			fmt.Println("LoadMapWithAPI() tried to insert something other than a string in Daemon.API")
		}
	}
}

func GetRealmAddress(region, slug, token string) (string, bool) {
	return "", true
}
func (d *Daemon) GetRealms() (Realms, bool) {
	// Build the address to get the realm index
	url, check := d.BuildRealmIndexAddress()
	if !check {
		return Realms{}, false
	}
	realms, check := d.CallRealmIndexAPI(url)
	if !check {
		return Realms{}, false
	}
	d.Realms = realms.Realms
	d.BuildAuctionURLS()
	return realms, true
}
func (d *Daemon) BuildAuctionURLS() {
	for i, v := range d.Realms {
		url, check := d.GetAPIStrings("api", "auctionrequest")
		if !check {
			fmt.Println("Error in BuildAuctionURLs()")
		}
		url = strings.Replace(url, regionString, d.Region, 1)
		url = strings.Replace(url, localeString, d.Locale, 1)
		url = strings.Replace(url, tokenString, d.Token.Token(), 1)
		url = strings.Replace(url, "{slug}", v.Slug, 1)

		d.Realms[i].URL = url
	}
}
func (d *Daemon) BuildRealmAddress(slug string) (string, bool) {
	realmAddress, check := d.GetAPIStrings("api", "realm")
	if !check {
		return "", false
	}
	realmAddress = strings.Replace(realmAddress, regionString, d.Region, 1)
	realmAddress = strings.Replace(realmAddress, localeString, d.Locale, 1)
	realmAddress = strings.Replace(realmAddress, tokenString, d.Token.Token(), 1)
	realmAddress = strings.Replace(realmAddress, "{slug}", slug, 1)
	return realmAddress, true
}

func (d *Daemon) BuildRealmIndexAddress() (string, bool) {
	// This string will contain 3 items that need replacement
	realmIndex, check := d.GetAPIStrings("api", "realmsIndex")
	if !check {
		return "", false
	}
	out := strings.Replace(realmIndex, regionString, d.Region, 1)
	out = strings.Replace(out, localeString, d.Locale, 1)
	out = strings.Replace(out, tokenString, d.Token.Token(), 1)
	return out, true
}

func CallRealmAPI(address string) (RealmData, bool) {
	client := http.Client{Timeout: 20 * time.Second}
	request, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return RealmData{}, true
	}
	res, err := client.Do(request)
	if err != nil {
		return RealmData{}, true
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return RealmData{}, true
	}

	rd := RealmData{}
	err = json.Unmarshal(body, &rd)
	if err != nil {
		fmt.Println("CallRealmAPI() generated an error in json.Unmarshal")
		return RealmData{}, true
	}
	return rd, false
}

func (d *Daemon) CallRealmIndexAPI(address string) (Realms, bool) {
	client := http.Client{Timeout: 20 * time.Second}
	request, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() failed using http.NewRequest()")
		fmt.Println(err)
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() failed using client.Do()")
		log.Fatal(err)
		return Realms{}, false
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() failed using ioutil.ReadAll()")
		return Realms{}, false
	}

	rd := Realms{}
	err = json.Unmarshal(body, &rd)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() generated an error in json.Unmarshal")
		fmt.Println(err.Error())
		return Realms{}, false
	}
	return rd, true
}

// GetAPIStrings gets all strings stored in the JSON, concatenates them in order and returns the string
func (d *Daemon) GetAPIStrings(components ...string) (string, bool) {
	var out string
	for _, v := range components {
		word, ok := d.API[v]
		if ok {
			out = out + word
		} else {
			fmt.Println("Error encountered in Daemon.GetAPIStrings()")
			return "", false
		}
	}
	return out, true
}

func InitializeDatabase() {
	// Create Realm tables
	// Create Item Table
	d := Daemon{Region: "us", Locale: "en_US"}
	d.LoadMapWithAPI()
	d.Token = NewToken()
	realms, ok := d.GetRealms()
	if !ok {
		log.Fatal("Could not create initialize the database")
	}
	d.Realms = realms.Realms
	dbInfo, ok := GetDBInfo()
	if !ok {
		log.Fatal("Could not create initialize the database")
	}
	db, ok := OpenDB(dbInfo)
	if !ok {
		log.Fatal("Could not create initialize the database")
	}
	for _, v := range realms.Realms {
		realmString := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"(id int8 primary key, item int4, owner varchar(50), orealm varchar(50), bid int8, buyout int8, quantity smallint, timeleft varchar(50), created timestamp with time zone);`,
			v.Slug)
		db.Exec(realmString)
	}
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS items(item int4 primary key, href varchar(150), name varchar(150));")
	check(err)
	statement.Exec()
	db.Close()
}
func (d *Daemon) monitor() {
	tick := time.Tick(time.Second * 10)
	for {
		<-tick
		for _, v := range d.AuctionManager {
			if len(v.Auctions) > 0 {
				fmt.Println(v.Realm.Slug + " - len(toAdd) " + strconv.Itoa(len(v.Auctions)))
			}
		}
	}
}

func (d *Daemon) auctionWorker() {

}
