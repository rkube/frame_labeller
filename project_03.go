package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
)

/*
 * some gemini stuff here *****
 * ignore for now *****s
type session_key struct {
	ID       uint16
	username string
}

var store = sessions.NewCookieStore([]byte("super-secret"))

func my_handler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context
	session := r.Context().Value(session_key{}).(*sessions.Session)

	// Access or modify session data
	username := session.Values["username"]
	fmt.Println("username = ", username)
	session.Values["counter"] = session.Values["counter"].(int) + 1

	// Save the session changes
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle the HTTP request based on the method and path
	switch r.Method {
	case "GET":
		// Handle GET requests
		// ...
	case "POST":
		// Handle POST requests
		// ...
	default:
		// Handle other methods or return an error
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func SessionMiddleware(store sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get or create a new session
			session, err := store.Get(r, "your_session_name")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Attach the session to the context
			ctx := context.WithValue(r.Context(), session_key{ID: 1, username: "test"}, session)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
*/

var users = map[string]string{
	"user1": "password-1",
	"user2": "password-2",
}

// this map stores the users sessions. For larger scale applications, you can use a database or cache for this purpose
var sessions = map[string]session{}

// each session contains the username of the user and the time at which it expires
type session struct {
	username string
	expiry   time.Time
}

// we'll use this method later to determine if the session has expired
func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

// Create a struct that models the structure of a user in the request body
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

func signin_handler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the expected password from our in memory map
	expectedPassword, ok := users[creds.Username]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create a new random session token
	// we use the "github.com/google/uuid" library to generate UUIDs
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	// Set the token in the session map, along with the session information
	sessions[sessionToken] = session{
		username: creds.Username,
		expiry:   expiresAt,
	}

	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 seconds
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})
}

// func middleware(handler http.Handler) http.Handler {
// 	return http.HandlerFunc(
// 		func(w http.ResponseWriter, r *http.Request) {
// 			fmt.Println("Executing middleware before request phase!")
// 			// Pass control back to the handler
// 			handler.ServeHTTP(w, r)
// 			fmt.Println("Executing middleware after response phase!")
// 		})
// }

type state_data struct {
	Username string
	Id       int
}

func my_route(w http.ResponseWriter, r *http.Request) {
	// Handling my route
	fmt.Println("Handling HTTP requests...")

	t_data := state_data{Username: "anonymous", Id: 0}

	// Read template
	templ, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("t_data = ", t_data)

	templ.Execute(w, t_data)

	// message := r.PostFormValue("message")
	// fmt.Println("messsage = ", message)
	// templ := template.Must(template.ParseFiles("index.html"))
	// todo := Todo{Id: len(data["Todos"]) + 1, Message: message}
	// data["Todos"] = append(data["Todos"], todo)

	// fmt.Println("todo: ", todo)
	// templ.ExecuteTemplate(w, "todo-list-element", todo)
}

func main() {
	fmt.Println("Hello, world")

	// Create a router or handler
	// router := http.NewServeMux()

	// Hook up the session middleware
	// router.Handle("/", (store)(http.HandlerFunc(my_handler)))

	route_handler := http.HandlerFunc(my_route)
	// http.Handle("/", middleware(route_handler))
	http.Handle("/", route_handler)
	http.Handle("/signin", http.HandlerFunc(signin_handler))

	// Start the server
	http.ListenAndServe(":8080", nil)
}
