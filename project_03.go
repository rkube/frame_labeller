package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shamaton/msgpack/v2"
)

// this map stores the users sessions. For larger scale applications, you can use a database or cache for this purpose
var sessions = map[string]session{}

// each session contains the username of the user and the time at which it expires
type session struct {
	expiry time.Time
}

/*
 * Generate a new session token
 */
func signin_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("signin_handler here")

	// Create a new random session token
	// we use the "github.com/google/uuid" library to generate UUIDs
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	// // Set the token in the session map, along with the session information
	sessions[sessionToken] = session{
		// username: creds.Username,
		// username: _username,
		expiry: expiresAt,
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 seconds
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})

	fmt.Fprintf(w, "setting new session token: %s", sessionToken)
}

type state_data struct {
	Session_token_id string
	// Id       int
}

func my_route(w http.ResponseWriter, r *http.Request) {
	// Handling my route
	fmt.Println("Handling HTTP requests...")

	t_data := state_data{Session_token_id: "undefined"}
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("Session token not set")
	} else {
		fmt.Println("session_token = ", c.Value)
		t_data.Session_token_id = c.Value
	}

	// Read template
	templ, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templ.Execute(w, t_data)
}

// Code below is for the API that returns data
type MyData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func fetch_data(w http.ResponseWriter, r *http.Request) {
	data := MyData{
		Name: "Alice",
		Age:  rand.Int(),
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error marshalling data", http.StatusInternalServerError)
		return
	}

	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON  data to the response writer
	w.Write(jsonData)
}

// This function sends a 100x100 16-bit integer array
func fetch_data_array(w http.ResponseWriter, r *http.Request) {
	// Create a 10x10 array of 16-bit integers
	array := make([]uint16, 100)
	for i := range array {
		array[i] = uint16(i)
		// array[i] = make([]uint16, 10)
		// for j := range array[i] {
		// 	array[i][j] = uint16(i*10 + j)
		// }
	}
	fmt.Println("Sending array data")

	// Serialize the array using MessagePack
	w.Header().Set("Content-Type", "application/x-msgpack")
	if err := msgpack.MarshalWriteAsArray(w, array); err != nil {
		panic(err)
	}

}
func main() {
	fmt.Println("Hello, world")

	http.Handle("/", http.HandlerFunc(my_route))
	http.Handle("/signin", http.HandlerFunc(signin_handler))
	http.Handle("/api/fetch_data", http.HandlerFunc(fetch_data))
	http.Handle("/api/fetch_data_uint16", http.HandlerFunc(fetch_data_array))

	// Start the server
	http.ListenAndServe(":8080", nil)
}
