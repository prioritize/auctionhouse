package auctionhouse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const timeout = 10

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
func NewToken() Token {
	credentials := getCredentials()
	t := Token{}
	t.Client = credentials.Client
	t.Secret = credentials.Secret
	t.token = ""
	// t.LastUpdated = time.Now()
	t.tokenURL = buildTokenURL("us")
	t.checkTokenURL = buildCheckURL("us")
	t.ValidateToken()
	return t
}
func getCredentials() Credentials {
	credentials := Credentials{}
	file, err := os.Open("../auctionjson/credentials.json")
	check(err)
	body, err := ioutil.ReadAll(file)
	check(err)
	err = json.Unmarshal(body, &credentials)
	return credentials
}

func (t *Token) ValidateToken() {
	if !t.checkValid() {
		t.updateToken()
	}
}

func (t *Token) Token() string {
	t.ValidateToken()
	return t.token
}
func (t *Token) checkValid() bool {
	if time.Since(t.LastUpdated) < time.Minute*5 {
		return true
	}
	holder := make(map[string]interface{}, 0)
	client := http.Client{Timeout: timeout * time.Second}
	supplied := "token=" + t.token
	fmt.Println(supplied)
	request, err := http.NewRequest(http.MethodPost, t.checkTokenURL, bytes.NewBuffer([]byte(supplied)))
	check(err)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	check(err)
	if response.StatusCode != 200 {
		return false
	}
	body, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(body, &holder)
	check(err)
	t.LastUpdated = time.Now()
	return true
}

func (t *Token) updateToken() {
	// get a new token
	// store the new token in t.token
	holder := make(map[string]interface{}, 0)
	client := http.Client{Timeout: timeout * time.Second}
	grantString := "grant_type=client_credentials"
	request, err := http.NewRequest(http.MethodPost, t.tokenURL, bytes.NewBuffer([]byte(grantString)))
	check(err)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(t.Client, t.Secret)
	response, err := client.Do(request)
	if response.StatusCode != 200 {
		fmt.Println("Token.updateToken() failed to get new token")
		return
	}
	check(err)
	body, err := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(body, &holder)
	test, ok := holder["access_token"]
	if ok {
		attempt, ok := test.(string)
		if ok {
			t.token = attempt
		}
	}
	fmt.Println(t.token)
}
func buildCheckURL(region string) string {
	url := "https://{region}.battle.net/oauth/check_token"
	out := strings.Replace(url, "{region}", region, 1)
	return out
}
func buildTokenURL(region string) string {
	url := "https://{region}.battle.net/oauth/token"
	out := strings.Replace(url, "{region}", region, 1)
	return out
}
