package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func root(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		fmt.Fprintf(w, "Sorry, there's some weird mumbo jumbo happening around here...")
	} else if auth, set := session.Values["auth"]; set {
		fmt.Fprintf(w, "<h1>Welcome, %v</h1>", auth)
	} else {
		fmt.Fprintf(w, "<h1>You're not logged in!</h1>")
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(r, "session-name")
	// Set some session values.
	session.Values["auth"] = 12345
	// Save it.
	session.Save(r, w)
}

func main() {
	http.HandleFunc("/", root)
	http.HandleFunc("/login", login)

	fmt.Println("Serving webserver...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
