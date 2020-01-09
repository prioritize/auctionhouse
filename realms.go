package auctionhouse

import (
	"database/sql"
	"fmt"
	"log"
)

func (r *Realm) fillAuctionMap(db *sql.DB) map[int]int {
	storage := make(map[int]int, 0)
	query := db
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
