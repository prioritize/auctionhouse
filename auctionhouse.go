package auctionhouse

import (
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

func GetRealmAddress(region, slug, token string) (string, bool) {
	var result map[string]interface{}
	base, err := os.Open("../auctionjson/api.json")
	if err != nil {
		return "", true
	}
	body, err := ioutil.ReadAll(base)
	if err != nil {
		return "", true
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", true
	}
	baseAddress, ok := result["apidata_base"]
	if !ok {
		return "", true
	}
	suffix, ok := result["realm"]
	if !ok {
		return "", true
	}
	suffixString, suffixOk := suffix.(string)
	baseAddressString, baseOk := baseAddress.(string)
	if suffixOk && baseOk {
		out := (BuildRealmAddress("us", baseAddressString, suffixString, slug, token))
		return out, false
	}
	return "", true
}
func GetRealms() ([]Realm, bool) {
	// Get the Realm Index
	// For each entry in the realm index query the realm page
	// Store the realm data into the database
	realms := make([]Realm, 0)
	// url := BuildIndexAddress()
	strings, check := GetAddress("apidata_base", "realmsIndex")
	if check {
		fmt.Println("GetRealms() failed using GetAddress()")
		return []Realm{}, true
	}
	address := BuildIndexAddress("us", strings[0]+strings[1])
	realms, check = CallRealmIndexAPI(address)
	if check {
		fmt.Println("GetRealms() failed using CallRealmIndexAPI()")
		return []Realm{}, true
	}
	return realms, false
}
func BuildRealmAddress(region, apiString, suffix, slug, token string) string {
	locale := "en_US"
	out := strings.Replace(apiString, "{region}", region, 1)
	out = out + strings.Replace(suffix, "{slug}", slug, 1)
	out = strings.Replace(out, "{locale}", locale, 1)
	return strings.Replace(out, "{token}", token, 1)

}
func BuildIndexAddress(region, apiString string) string {
	return strings.Replace(apiString, "{region}", region, 1)
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

func CallRealmIndexAPI(address string) ([]Realm, bool) {
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
		return []Realm{}, true
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(res.StatusCode)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() failed using ioutil.ReadAll()")
		return []Realm{}, true
	}

	rd := []Realm{}
	fmt.Println(string(body))
	err = json.Unmarshal(body, &rd)
	if err != nil {
		fmt.Println("CallRealmIndexAPI() generated an error in json.Unmarshal")
		return []Realm{}, true
	}
	return rd, false
}
