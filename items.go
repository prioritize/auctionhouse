package auctionhouse

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func NewItemManager() ItemManager {
	i := ItemManager{}
	i.toAdd = make(chan Item, 500)
	i.Items = make(map[int]int, 0)
	i.toQueryAPI = make(chan toQueryAPI, 500)
	i.fillDBInfo()
	i.openDB()
	return i
}
// !! This is not complete, and need to outline how this function should work
// !! This should likely just get all information from the database and handle 
// !! querying the API to another function
func (i *ItemManager) populateItemManager() {
	query, err := i.db.Prepare("SELECT DISTINCT item from items;")
	check(err)
	// Get all of the item numbers from the database
	var queryResponse []string
	rows, err := query.Query()
	check(err)
	defer rows.Close()
	// Store all of those item numbers from the database into the queryresponse slice
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		i.Items[]
	}

	for _, v := range queryResponse {
		value, err := strconv.Atoi(v)
		if err == nil {
			i.Items[value] = 0
		}
	}
}
func (i *ItemManager) fillDBInfo() {
	file, err := os.Open("../auctionjson/database.jon")
	check(err)
	body, err := ioutil.ReadAll(file)
	check(err)
	dbInfo := DBInfo{}
	err = json.Unmarshal(body, &dbInfo)
	check(err)
	i.dbInfo = dbInfo
}
func (i *ItemManager) openDB() {
	psqlConnInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		i.dbInfo.Host, i.dbInfo.Port, i.dbInfo.User, i.dbInfo.Password, i.dbInfo.DBname)
	database, err := sql.Open("postgres", psqlConnInfo)
	check(err)
	i.db = database
}
