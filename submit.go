package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

/*
 * Stores label and comment in the database
 */
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
	fmt.Println("/api/handle_submit: shotnr = ", a.all_user_state[username].Sparta_state.Shotnr)
	fmt.Println("/api/handle_submit: framen = ", a.all_user_state[username].Sparta_state.Current_frame)
	fmt.Println("/api/handle_submit: label = ", r.Form["label"][0])
	fmt.Println("/api/handle_submit: comment = ", r.Form["comment"][0])

	// // Prepare SQL command
	sql_cmd := "INSERT INTO test01 (username, shotnr, framenr, label, comment) VALUES "
	sql_cmd += fmt.Sprintf("(\"%s\", ", username)
	sql_cmd += fmt.Sprintf("%d, ", a.all_user_state[username].Sparta_state.Shotnr)
	sql_cmd += fmt.Sprintf("%d, ", a.all_user_state[username].Sparta_state.Current_frame)
	sql_cmd += fmt.Sprintf("\"%s\", ", r.Form["label"][0])
	sql_cmd += fmt.Sprintf("\"%s\")", r.Form["comment"][0])
	//  strconv.Itoa(a.all_user_state[username].shotnr) + ", "

	fmt.Println("/api/handle_submit: sql_cmd = ", sql_cmd)

	// onnect to database
	db, err := sql.Open("sqlite", "test.sql")
	if err != nil {
		fmt.Println("Could not connnect to database")
	}

	sql_res, err := db.Exec(sql_cmd)
	if err != nil {
		fmt.Println("Could not execute command: ", sql_cmd)
	}

	fmt.Println("res = ", sql_res)

	return 200, nil
}
