package main

import (
	"fmt"
	"net/http"

	"database/sql"

	_ "modernc.org/sqlite"
)

// Stores submitted label and comment in backend
func handle_submit(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handle_submit here")
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("handle_submit: Session token not set")
	} else {
		fmt.Println("handle_submit: session_token = ", c.Value)
	}

	// Connect to database
	db, err := sql.Open("sqlite", "test.sql")
	if err != nil {
		fmt.Println("Could not connnect to database")
	}

	rows, err := db.Query("select * from test01;")
	if err != nil {
		fmt.Println("Error during query: ", err)
	}

	// ...> "username" Text,
	// ...> "shotnr" Int,
	// ...> "framenr" Int,
	// ...> "label" Text,
	// ...> "comment" Text);
	for rows.Next() {
		var sql_username string
		var sql_shotnr int
		var sql_framenr int
		var sql_label string
		var sql_comment string

		if err = rows.Scan(&sql_username, &sql_shotnr, &sql_framenr, &sql_label, &sql_comment); err != nil {
			// return err
			fmt.Println("Error during scan: ", err)
		}

		fmt.Println("sql_user = ", sql_username)
		fmt.Println("sql_shotnr = ", sql_shotnr)
		fmt.Println("sql_framenr = ", sql_framenr)
		fmt.Println("sql_label = ", sql_label)
		fmt.Println("sql_comment = ", sql_comment)
	}

	fmt.Fprintf(w, "submitted data")
}
