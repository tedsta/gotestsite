package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello, world!</h1>")
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(r, "session-name")
	// Set some session values.
	session.Values["foo"] = "bar"
	session.Values[42] = 43
	// Save it.
	session.Save(r, w)
}

func main() {
	http.HandleFunc("/", root)
	http.HandleFunc("/login", handler)

	fmt.Println("Serving webserver...")
	http.ListenAndServe("", nil)
}
