package auctionhouse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Realm struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}
type RealmData struct {
	ID           int    `json:"id"`
	Region       Region `json:"region"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Locale       string `json:"locale"`
	Timezone     string `json:"timezone"`
	Type         string `json:"type"`
	IsTournament bool   `json:"is_tournament"`
	Slug         string `json:"slug"`
}

type Region struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
type RealmType struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func GetRealmAddress(region string) (string, bool) {
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
	fmt.Println(result)
	baseAddress, ok := result["apidata_base"]
	t, ok := baseAddress.(string)
	if ok {
		fmt.Println(ReplaceRegion("us", t))
	}
	return "1", true
}
func ReplaceRegion(region, apiString string) string {
	return strings.Replace(apiString, "{region}", "us", 1)
}

// func GetRealms() []Realm {

// }
