package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Keeps track where each user is at a given time
type user_state struct {
	shotnr             int32
	frame              int32
	current_session_id string
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
	// Initialize with a blank state.
	this_state := user_state{shotnr: 0, frame: 1, current_session_id: ""}
	// If the request comes with a session cookie, we can recover the previous state
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("/my_route: Session token not set")
	} else {
		fmt.Println("/my_route: session_token = ", c.Value)
		if username, ok := a.session_to_user[c.Value]; ok {
			fmt.Println("/my_route: found username associated with sesion id= ", username)
			this_state.current_session_id = c.Value
			this_state.frame = a.all_user_state[username].frame
			this_state.shotnr = a.all_user_state[username].shotnr
		}
	}

	fmt.Println("/my_route: Using state: ", this_state)

	t, err := template.ParseFiles("templates/main.tmpl", "templates/signin.tmpl")
	if err != nil {
		fmt.Println("Error parsing files")
	}

	err = t.Execute(w, a)
	if err != nil {
		fmt.Println("Error executing templates")
	}

	return 0, nil
}

/*
 * Generate a new session token
 */
func signin_handler(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	new_state := user_state{shotnr: 0, frame: 1, current_session_id: ""}
	// Extract username from form.
	err := r.ParseForm()
	if err != nil { // Return a Bad Request if we can't parse the form
		fmt.Fprintf(os.Stdout, "/api/signin_handler: Unable to parse %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return 400, nil
	}
	username := r.Form.Get("username")
	fmt.Printf("/api/signin_handler: username = %s\n", username)

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
		SameSite: http.SameSiteLaxMode,
	})

	new_state.current_session_id = sessionToken

	// Handle the case where the user has an old session id registered with him
	if old_user_state, ok := a.all_user_state[username]; ok {
		// If the user has a previously assigned session-id
		// - update the mapping from session id to user.
		delete(a.session_to_user, old_user_state.current_session_id)
		// Insert the new session token for the user, but keep the rest of the state the same
		new_state.shotnr = old_user_state.shotnr
		new_state.frame = old_user_state.frame
	} // Nothing to do if the user registerd for the first time
	a.session_to_user[sessionToken] = username
	a.all_user_state[username] = new_state

	fmt.Println("/api/signin_handler: a = ", a)

	// Write a response, this will be rendered by htmx
	fmt.Fprintf(w, "setting new session token: %s", sessionToken)
	fmt.Printf("/api/signin_handler: Setting new session token: %s\n", sessionToken)
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
// Timing, number of frames, Geometry: TBD
// Also loads sparta frame navigation panel
func get_sparta_info(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	// This route will update the state of a user.
	new_state := user_state{shotnr: 0, frame: 1, current_session_id: ""}

	// Parse the passed form, extract shotnr, and update state.
	// Throw an error if shotnr is not passed
	err := r.ParseForm()
	if err != nil { // Return a Bad Request if we can't parse the form
		fmt.Fprintf(os.Stdout, "get_sparta_info: Unable to parse %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return 400, nil
	}
	// Extract shotnr, convert to int and update state
	_shotnr, ok := r.Form["shotnr"]
	if ok {
		_s, err := strconv.Atoi(_shotnr[0])
		if err != nil {
			err_msg := fmt.Sprintf("/get_spart_info: Could not parse %s as int.", _shotnr[0])
			w.WriteHeader(http.StatusBadRequest)
			return -1, errors.New(err_msg)
		} else {
			new_state.shotnr = int32(_s)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return -1, errors.New("route /get_spart_info requires a shot number")
	}

	// If the request came with a session token, update the state for the user
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("/get_sparta_info: Session token not set")
	} else {
		fmt.Println("/get_sparta_info: session_token = ", c.Value)
		username := a.session_to_user[c.Value]
		fmt.Println("/get_sparta_info: username = ", username)
		// The user should have already been linked to his id. Verify this.
		fmt.Println("/get_sparta_info: registered sesssion-id: ", a.all_user_state[username].current_session_id)
		fmt.Println("/get_sparta_info: session-id in request: ", c.Value)

		if a.all_user_state[username].current_session_id != c.Value {
			fmt.Fprintf(os.Stdout, "/get_sparta_info: username %s has registered session-id %s, but provided %s\n", username, a.all_user_state[username].current_session_id, c.Value)
		}
		// Update state
		new_state.current_session_id = c.Value
		a.all_user_state[username] = new_state
		fmt.Println("New state: ", a.all_user_state[username])
	}

	my_sparta_info := sparta_info{Shotnr: int(new_state.shotnr), T_start: 0.0, T_end: 1.0, Num_frames: rand.Int() % 100}
	tmpl := template.Must(template.ParseFiles("templates/sparta_info.tmpl"))
	tmpl.Execute(w, my_sparta_info)

	return 0, nil
}

func main() {
	_, err := os.Stat(filepath.Join(".", "css", "style.css"))
	if err != nil {
		println(err.Error())
	}

	context := &app_context{make(map[string]string), make(map[string]user_state)}

	http.Handle("/", app_handler{context, my_route})
	http.Handle("/api/signin", app_handler{context, signin_handler})           // Handles login etc.
	http.Handle("/api/get_sparta_info", app_handler{context, get_sparta_info}) // Loads SPARTA info
	http.Handle("/api/submit", app_handler{context, handle_submit})            // Handles label submission etc.
	http.Handle("/api/sparta_frame", app_handler{context, fetch_sparta_plot})
	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("css/")))) // to serve css

	// Start the server
	http.ListenAndServe(":8080", nil)
}
