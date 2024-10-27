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
	Username           string
	Current_session_id string
	Sparta_state       sparta_info
}

// Structure to store sparta information for a given shot
// Remember to capitalize member names so that they are public
// https://dev.to/jpoly1219/structs-methods-and-receivers-in-go-5g4f
type sparta_info struct {
	Shotnr        int
	T_start       float64
	T_end         float64
	Num_frames    int
	Current_frame int
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
	//
	new_state := user_state{Current_session_id: "",
		Username:     "anon",
		Sparta_state: sparta_info{Shotnr: 0, T_start: -1.0, T_end: -1.0, Num_frames: 0, Current_frame: 0}}
	// If the request comes with a session cookie, we can recover the previous state
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println("/my_route: Session token not set")
	} else {
		fmt.Println("/my_route: session_token = ", c.Value)
		if username, ok := a.session_to_user[c.Value]; ok {
			new_state.Username = username
			new_state.Current_session_id = c.Value
			new_state.Sparta_state = a.all_user_state[username].Sparta_state
		}
	}

	// Render templates
	fmt.Println("/my_route: Using state: ", new_state)
	t, err := template.ParseFiles("templates/main.tmpl", "templates/signin.tmpl", "templates/header.tmpl", "templates/sparta_info.tmpl", "templates/frame_navigation.tmpl")
	if err != nil {
		fmt.Println("Error parsing files")
	}

	err = t.Execute(w, new_state)
	if err != nil {
		fmt.Println("/my_route: Error executing templates")
	}

	return 0, nil
}

/*
 * Handles user login requests.
 * Parameters: username (from form)
 * Generate a new session token and associate it with the user in the app_context.
 * If the user was previously identified with a different session token, replace it with the
 * newly generated.
 * Return a cookie to the client with the session_token.
 * Also, write a html fragment to update the header.
 */
func signin_handler(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Println("/api/signin_handler: starting")
	new_state := user_state{Current_session_id: "",
		Username:     "anon",
		Sparta_state: sparta_info{Shotnr: 0, T_start: -1.0, T_end: -1.0, Num_frames: 0, Current_frame: 0}}

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
	// use the "github.com/google/uuid" library to generate UUIDs
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

	new_state.Current_session_id = sessionToken

	// Handle the case where the user has an old session id registered with him
	if old_user_state, ok := a.all_user_state[username]; ok {
		// If the user has a previously assigned session-id
		// - update the mapping from session id to user.
		delete(a.session_to_user, old_user_state.Current_session_id)
		// Insert the new session token for the user, but keep the rest of the state the same
		new_state.Sparta_state = old_user_state.Sparta_state
		// new_state.Shotnr = old_user_state.Shotnr
		// new_state.Frame = old_user_state.Frame
	} // Nothing to do if the user registerd for the first time
	a.session_to_user[sessionToken] = username
	a.all_user_state[username] = new_state

	fmt.Println("/api/signin_handler: a = ", a)

	// Update the header to show the new username
	// tmpl_str := string("")
	// tmpl_bytes, err := os.ReadFile("templates/header.tmpl")
	// if err != nil {
	// 	fmt.Println("/api/signin_handler: Could not read template from templates/header.tmpl")
	// }

	// tmpl_str += string(tmpl_bytes)
	// tmpl_str += ` {{ template "page_header_tmpl" . }}`
	// tmpl := template.Must(template.New("").Parse(tmpl_str))
	// tmpl.Execute(w, new_state)

	// Write a response, this will be rendered by htmx
	fmt.Fprintf(w, "setting new session token: %s", sessionToken)
	fmt.Printf("/api/signin_handler: Setting new session token: %s\n", sessionToken)
	return 0, nil
}

// Loads information on SPARTA for a given shot
// Timing, number of frames, Geometry: TBD
// Also loads sparta frame navigation panel
func get_sparta_info(a *app_context, w http.ResponseWriter, r *http.Request) (int, error) {
	// This route will update the state of a user:
	// set user_state.Sparta_info to
	new_state := user_state{}

	// Parse the passed form, extract shotnr, and update state.
	// Throw an error if shotnr is not passed
	err := r.ParseForm()
	if err != nil { // Return a Bad Request if we can't parse the form
		fmt.Fprintf(os.Stdout, "get_sparta_info: Unable to parse %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return 400, nil
	}
	// Extract shotnr, convert to int, and instantiate new sparta state
	_shotnr, ok := r.Form["shotnr"]
	if ok {
		_s, err := strconv.Atoi(_shotnr[0])
		if err != nil {
			err_msg := fmt.Sprintf("/get_spart_info: Could not parse %s as int.", _shotnr[0])
			w.WriteHeader(http.StatusBadRequest)
			return -1, errors.New(err_msg)
		} else {
			new_state.Sparta_state = sparta_info{Shotnr: int(_s), T_start: 0.0, T_end: 1.0, Num_frames: rand.Int() % 100}
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
		fmt.Println("/get_sparta_info: registered sesssion-id: ", a.all_user_state[username].Current_session_id)
		fmt.Println("/get_sparta_info: session-id in request: ", c.Value)

		if a.all_user_state[username].Current_session_id != c.Value {
			fmt.Fprintf(os.Stdout, "/get_sparta_info: username %s has registered session-id %s, but provided %s\n", username, a.all_user_state[username].Current_session_id, c.Value)
		}
		// Update state
		new_state.Current_session_id = c.Value
		new_state.Username = username
		a.all_user_state[username] = new_state
		fmt.Println("New state: ", a.all_user_state[username])
	}

	// Read all template definition files and instantiate them
	filenames := [2]string{"templates/sparta_info.tmpl", "templates/frame_navigation.tmpl"}
	tmpl_names := [2]string{"shot_info_tmpl", "framenav_tmpl"}
	tmpl_str := string("")

	// Read all template definitions from file into a string.
	// For each template definition, add a string that instantiates that template
	for ix, filename := range filenames {
		tmpl_bytes, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println("Could not read from file ", filename)
		}
		tmpl_str += string(tmpl_bytes)
		tmpl_str += ` {{ template "` + tmpl_names[ix] + `" . }}`
		if ix < len(filenames) {
			tmpl_str += "\n\n"
		}
	}
	tmpl := template.Must(template.New("").Parse(tmpl_str))
	tmpl.Execute(w, new_state)

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
