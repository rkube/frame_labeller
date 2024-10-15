package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/shamaton/msgpack/v2"
)

// Keeps track where each user is at a given time
type user_state struct {
	username string
	shotnr   uint
	frame    uint
}

// app_context is the local context. It is created in main and passed to all http handlers.
type app_context struct {
	session_to_user map[string]string     // Map from session ids to user ids
	all_user_state  map[string]user_state // Map from user id to user_state

}

// // this map stores the users sessions. For larger scale applications, you can use a database or cache for this purpose
// var sessions = map[string]session{}

// // each session contains the username of the user and the time at which it expires
// type session struct {
// 	expiry time.Time
// }

// Maps session-ids to users
// var session_to_user = map[string]user_id{}

// type user_id struct {
// 	name string
// }

// // Maps usernames to the current state
// var state_map = map[user_id]state{}

/*
 * Generate a new session token
 */
func signin_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("signin_handler here")

	// Extract username from form.
	err := r.ParseForm()
	if err != nil { // Return a Bad Request if we can't parse the form
		fmt.Fprintf(os.Stdout, "signin_handler: Unable to parse %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	fmt.Println("r.Form = ", r.Form)
	username := r.Form.Get("username")
	fmt.Printf("signin_handler: username = %s\n", username)

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
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: false,
		Path:     "/",
		// Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Write a response, this will be rendered by htmx
	fmt.Fprintf(w, "setting new session token: %s", sessionToken)
	fmt.Printf("signin_handler: Setting new session token: %s\n", sessionToken)
}

// Tracks users with a unique session string
// type state_data struct {
// 	Session_token_id string
// }

// Renders the main page
func my_route(w http.ResponseWriter, r *http.Request) {
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

// Sends SPARTA dummy data (100x100 16-bit integer array, random values)
func fetch_data_array(w http.ResponseWriter, r *http.Request) {
	t_data := state_data{Session_token_id: "undefined"}

	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("fetch_data_array: Session token not set")
	} else {
		fmt.Println("fetch_data_array: session_token = ", c.Value)
		t_data.Session_token_id = c.Value
	}

	// Create a 10x10 array of 16-bit integers
	array := make([]uint16, 100)
	for i := range array {
		array[i] = uint16(rand.Int() % 100)
	}
	fmt.Println("fetch_data_array: Sending array data")

	// Serialize the array using MessagePack
	w.Header().Set("Content-Type", "application/x-msgpack")
	if err := msgpack.MarshalWriteAsArray(w, array); err != nil {
		panic(err)
	}

}

// Structure to store sparta information for a given shot
// Remember to capitalize member names so that they are public
// https://dev.to/jpoly1219/structs-methods-and-receivers-in-go-5g4f
type sparta_info struct {
	Shotnr     int
	T_start    float64
	T_end      float64
	Num_frames int
}

// Loads information on SPARTA for a given shot
// Geometry - TBD
// Timing s
func get_sparta_info(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("get_sparta_info: Session token not set")
	} else {
		fmt.Println("get_sparta_info: session_token = ", c.Value)

	}
	my_sparta_info := sparta_info{Shotnr: 241013010, T_start: 0.0, T_end: 0.0, Num_frames: rand.Int() % 100}
	tmpl := template.Must(template.ParseFiles("templates/sparta_info.tmpl"))
	tmpl.Execute(w, my_sparta_info)
}

// Stores submitted label and comment in backend
func handle_submit(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handle_submit here")
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("handle_submit: Session token not set")
	} else {
		fmt.Println("handle_submit: session_token = ", c.Value)
	}

	fmt.Fprintf(w, "submitted data")
}

// Creates frame navigation slider
type frame_nav_info struct {
	Num_frames    int
	Current_frame int
}

func fetch_frame_navigation(w http.ResponseWriter, r *http.Request) {
	fmt.Println("rendering frame navigation")
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("fetch_frame_navigation: Session token not set")
	} else {
		fmt.Println("fetch_frame_navigation: session_token = ", c.Value)
	}

	fi := frame_nav_info{Num_frames: 100, Current_frame: 1}
	tmpl := template.Must(template.ParseFiles("templates/frame_navigation.tmpl"))
	tmpl.Execute(w, fi)
}

func main() {
	fmt.Println("Hello, world")
	_, err := os.Stat(filepath.Join(".", "css", "style.css"))
	if err != nil {
		println(err.Error())
	}

	http.Handle("/", http.HandlerFunc(my_route))                                      // Main page
	http.Handle("/signin", http.HandlerFunc(signin_handler))                          // Handles login etc.
	http.Handle("/frame_navigation", http.HandlerFunc(fetch_frame_navigation))        // Loads frame navigation UI
	http.Handle("/get_sparta_info/", http.HandlerFunc(get_sparta_info))               // Loads SPARTA info
	http.Handle("/api/fetch_data_uint16", http.HandlerFunc(fetch_data_array))         // Loads SPARTA frames
	http.Handle("/api/submit", http.HandlerFunc(handle_submit))                       // Handles label submission etc.
	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("css/")))) // to serve css

	// Start the server
	http.ListenAndServe(":8080", nil)
}
