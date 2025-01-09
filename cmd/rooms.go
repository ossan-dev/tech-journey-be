package main

import (
	"context"
	"database/sql"
	"io"
)

func getRoomsIDs(roomsIDs *[]string) (err error) {
	db, err := sql.Open("postgres", "host=localhost port=54322 user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		return
	}
	err = db.Ping()
	if err != nil {
		return
	}
	ids, err := db.QueryContext(context.Background(), "SELECT id FROM public.rooms")
	if err != nil {
		return
	}
	for ids.Next() {
		var id string
		err = ids.Scan(&id)
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		*roomsIDs = append(*roomsIDs, id)
	}
	err = ids.Err()
	if err != nil {
		return
	}
	defer ids.Close()
	return
}
