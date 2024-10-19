package main

import (
	"fmt"
	"html/template"
	"log"
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
	shotnr uint
	frame  uint
}

// app_context is the local context. It is created in main and passed to all http handlers.
type app_context struct {
	session_to_user map[string]string     // Map from session ids to user ids
	all_user_state  map[string]user_state // Map from user id to user_state

}

/*
 * Make all handlers a struct that implements http.Handler
 * See https://drstearns.github.io/tutorials/gohandlerctx/
 * and https://blog.questionable.services/article/custom-handlers-avoiding-globals/
 */

type app_handler struct {
	*app_context
	H func(*app_context, http.ResponseWriter, *http.Request) (int, error)
}

func (ah app_handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Updated to pass ah.appContext as a parameter to our handler type.
	status, err := ah.H(ah.app_context, w, r)
	if err != nil {
		log.Printf("HTTP %d: %q", status, err)
	}
}

/*
 * Render the main page
 */
func my_route(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Println("Handling HTTP requests...")

	// t_data := state_data{Session_token_id: "undefined"}
	// session_id := ""
	// c, err := r.Cookie("session_token")
	// if err != nil {
	// 	fmt.Println("my_route: Session token not set")
	// 	session_id = "undefined"
	// } else {
	// 	fmt.Println("my_route: session_token = ", c.Value)
	// 	// t_data.Session_token_id = c.Value
	// 	session_id = c.Value
	// }

	http.ServeFile(w, r, "index.html")

	// Read template
	// templ, err := template.ParseFiles("index.html")
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return 0, err
	// }
	// templ.Execute(w, session_id)
	return 0, nil
}

/*
 * Generate a new session token
 */
func signin_handler(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Println("signin_handler here")

	// Extract username from form.
	err := r.ParseForm()
	if err != nil { // Return a Bad Request if we can't parse the form
		fmt.Fprintf(os.Stdout, "signin_handler: Unable to parse %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return 400, nil
	}
	username := r.Form.Get("username")
	fmt.Printf("signin_handler: username = %s\n", username)

	// Create a new random session token
	// we use the "github.com/google/uuid" library to generate UUIDs
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

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

	// TODO: Find if another sessionid belongs to the user.
	// If so, copy the old user_state to a new entry and remove the old session id.

	// Track the current user by its session ids
	a.session_to_user[sessionToken] = username
	a.all_user_state[username] = user_state{shotnr: 0, frame: 0}

	fmt.Println("signin_handler: a = ", a)

	// Write a response, this will be rendered by htmx
	fmt.Fprintf(w, "setting new session token: %s", sessionToken)
	fmt.Printf("signin_handler: Setting new session token: %s\n", sessionToken)
	return 0, nil
}

// Sends SPARTA dummy data (100x100 16-bit integer array, random values)
func fetch_data_array(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	// t_data := state_data{Session_token_id: "undefined"}
	session_id := ""

	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("fetch_data_array: Session token not set")
	} else {
		session_id = c.Value
		fmt.Println("fetch_data_array: session_token = ", session_id)
	}

	fmt.Println("fetch_data_array: a = ", a)

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

	return 0, nil

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

	context := &app_context{make(map[string]string), make(map[string]user_state)}

	// http.Handle("/", http.HandlerFunc(my_route))                                      // Main page
	http.Handle("/", app_handler{context, my_route})
	http.Handle("/signin", app_handler{context, signin_handler})                      // Handles login etc.
	http.Handle("/frame_navigation", http.HandlerFunc(fetch_frame_navigation))        // Loads frame navigation UI
	http.Handle("/get_sparta_info/", http.HandlerFunc(get_sparta_info))               // Loads SPARTA info
	http.Handle("/api/fetch_data_uint16", app_handler{context, fetch_data_array})     // Loads SPARTA frames
	http.Handle("/api/submit", http.HandlerFunc(handle_submit))                       // Handles label submission etc.
	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("css/")))) // to serve css

	// Start the server
	http.ListenAndServe(":8080", nil)
}
