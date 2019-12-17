package auctionhouse

import (
	"auctionauth"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type Realm struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}
type RealmData struct {
	ID           int       `json:"id"`
	Region       Region    `json:"region"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Locale       string    `json:"locale"`
	Timezone     string    `json:"timezone"`
	Type         RealmType `json:"type"`
	IsTournament bool      `json:"is_tournament"`
	Slug         string    `json:"slug"`
}

type Region struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}
type RealmType struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
type Daemon struct {
	Token  auctionauth.Token
	ID     int
	API    map[string]string
	Region string
	Locale string
	Realms []Realm
}

const regionString = "{region}"
const localeString = "{locale}"
const tokenString = "{token}"

func NewDaemon(region, locale string) (Daemon, bool) {
	d := Daemon{Region: region, Locale: locale}
	token, check := auctionauth.GetNewToken()
	if !check {
		d.Token = token
	} else {
		return Daemon{}, true
	}
	d.LoadMapWithAPI()
	for k, v := range d.API {
		fmt.Println(k, v)
	}
	return d, false
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
func GetAddress(labels ...string) ([]string, bool) {
	out := make([]string, 0)
	var addresses map[string]interface{}
	file, err := os.Open("../auctionjson/api.json")
	if err != nil {
		fmt.Println("GetAddress() failed using os.Open()")
		return []string{}, true
	}
	body, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("GetAddress() failed using ioutil.ReadAll()")
		return []string{}, true
	}
	err = json.Unmarshal(body, &addresses)
	if err != nil {
		fmt.Println("GetAddress() failed using json.Unmarshal()")
		return []string{}, true
	}
	for _, v := range labels {
		label, ok := addresses[v].(string)
		if ok {
			out = append(out, label)
		}
	}
	return out, false
}

// func GetRealmAddress(region, slug, token string) (string, bool) {
// 	var result map[string]interface{}
// 	base, err := os.Open("../auctionjson/api.json")
// 	if err != nil {
// 		return "", true
// 	}
// 	body, err := ioutil.ReadAll(base)
// 	if err != nil {
// 		return "", true
// 	}
// 	err = json.Unmarshal(body, &result)
// 	if err != nil {
// 		return "", true
// 	}
// 	baseAddress, ok := result["apidata_base"]
// 	if !ok {
// 		return "", true
// 	}
// 	suffix, ok := result["realm"]
// 	if !ok {
// 		return "", true
// 	}
// 	suffixString, suffixOk := suffix.(string)
// 	baseAddressString, baseOk := baseAddress.(string)
// 	if suffixOk && baseOk {
// 		out := (BuildRealmAddress("us", baseAddressString, suffixString, slug, token))
// 		return out, false
// 	}
// 	return "", true
// }
func (d *Daemon) GetRealms() ([]Realm, bool) {
	// Build the address to get the realm index
	url, check := d.BuildRealmIndexAddress()
	if !check {
		return []Realm{}, false
	}
	fmt.Println(url)
	// Get the Realm Index
	// For each entry in the realm index query the realm page
	// Store the realm data into the database
	realms, check := d.CallRealmIndexAPI(url)
	if !check {
		return []Realm{}, false
	}
	fmt.Println(realms)
	for _, v := range realms {
		fmt.Println(v.Slug)
	}

	return []Realm{}, false
}
func (d *Daemon) BuildRealmAddress(region, apiString, suffix, slug, token string) string {
	locale := "en_US"
	out := strings.Replace(apiString, "{region}", region, 1)
	out = out + strings.Replace(suffix, "{slug}", slug, 1)
	out = strings.Replace(out, "{locale}", locale, 1)
	return strings.Replace(out, "{token}", token, 1)

}

func (d *Daemon) BuildRealmIndexAddress() (string, bool) {
	// This string will contain 3 items that need replacement
	realmIndex, check := d.GetAPIStrings("api", "realmsIndex")
	if !check {
		return "", false
	}
	out := strings.Replace(realmIndex, regionString, d.Region, 1)
	out = strings.Replace(out, localeString, d.Locale, 1)
	out = strings.Replace(out, tokenString, d.Token.Token, 1)
	return out, true
}

func CallRealmAPI(address string) (RealmData, bool) {
	client := http.Client{Timeout: 5 * time.Second}
	fmt.Println(address)
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

func (d *Daemon) CallRealmIndexAPI(address string) ([]Realm, bool) {
	fmt.Println(address)
	client := http.Client{Timeout: 5 * time.Second}
	fmt.Println(address)
	request, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() failed using http.NewRequest()")
		fmt.Println(err)
	}
	fmt.Println("The address to client.Do() is " + address)
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() failed using client.Do()")
		return []Realm{}, false
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(res.StatusCode)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() failed using ioutil.ReadAll()")
		return []Realm{}, false
	}

	rd := []Realm{}
	fmt.Println(string(body))
	err = json.Unmarshal(body, &rd)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() generated an error in json.Unmarshal")
		return []Realm{}, false
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
