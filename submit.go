package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

// Stores submitted label and comment in backend
func handle_submit(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("/api/handle_submit: Session token not set")
	} else {
		fmt.Println("/api/handle_submit: session_token = ", c.Value)
	}

	// Parse form submission data
	err = r.ParseForm()
	if err != nil { // Return a Bad Request if we can't parse the form
		err_msg := fmt.Sprintf("/api/handle_submit: Unable to parse %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return 400, errors.New(err_msg)
	} else {
		_, label_ok := r.Form["label"]
		_, comment_ok := r.Form["comment"]
		if !(label_ok && comment_ok) {
			// Either label or comment are missing from the erquest.
			w.WriteHeader(http.StatusBadRequest)
			return 400, errors.New("request did not include both, label and comment")
		}
	}

	// Submit only happens for registerd users
	c, err = r.Cookie("session_token")
	if err != nil {
		fmt.Println("fetch_data_array: Session token not set")
	}

	username := a.session_to_user[c.Value]

	fmt.Println("/api/handle_submit: username = ", username)
	fmt.Println("/api/handle_submit: shotnr = ", a.all_user_state[username].shotnr)
	fmt.Println("/api/handle_submit: framen = ", a.all_user_state[username].frame)
	fmt.Println("/api/handle_submit: label = ", r.Form["label"][0])
	fmt.Println("/api/handle_submit: comment = ", r.Form["comment"][0])

	// // Prepare SQL command
	sql_cmd := "INSERT INTO test01 (username, shotnr, framenr, label, comment) VALUES "
	sql_cmd += fmt.Sprintf("(\"%s\", ", username)
	sql_cmd += fmt.Sprintf("%d, ", a.all_user_state[username].shotnr)
	sql_cmd += fmt.Sprintf("%d, ", a.all_user_state[username].frame)
	sql_cmd += fmt.Sprintf("\"%s\", ", r.Form["label"][0])
	sql_cmd += fmt.Sprintf("\"%s\")", r.Form["comment"][0])
	//  strconv.Itoa(a.all_user_state[username].shotnr) + ", "

	fmt.Println("/api/handle_submit: sql_cmd = ", sql_cmd)

	// // Connect to database
	db, err := sql.Open("sqlite", "test.sql")
	if err != nil {
		fmt.Println("Could not connnect to database")
	}

	sql_res, err := db.Exec(sql_cmd)
	if err != nil {
		fmt.Println("Could not execute command: ", sql_cmd)
	}

	fmt.Println("res = ", sql_res)

	// rows, err := db.Query("select * from test01;")
	// if err != nil {
	// 	fmt.Println("Error during query: ", err)
	// }

	// ...> "username" Text,
	// ...> "shotnr" Int,
	// ...> "framenr" Int,
	// ...> "label" Text,
	// ...> "comment" Text);
	// for rows.Next() {
	// 	var sql_username string
	// 	var sql_shotnr int
	// 	var sql_framenr int
	// 	var sql_label string
	// 	var sql_comment string

	// 	if err = rows.Scan(&sql_username, &sql_shotnr, &sql_framenr, &sql_label, &sql_comment); err != nil {
	// 		// return err
	// 		fmt.Println("Error during scan: ", err)
	// 	}

	// 	fmt.Println("sql_user = ", sql_username)
	// 	fmt.Println("sql_shotnr = ", sql_shotnr)
	// 	fmt.Println("sql_framenr = ", sql_framenr)
	// 	fmt.Println("sql_label = ", sql_label)
	// 	fmt.Println("sql_comment = ", sql_comment)
	// }

	// fmt.Fprintf(w, "submitted data")

	return 200, nil
}
