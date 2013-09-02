package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Context struct {
	db    *sql.DB
	store *sessions.CookieStore
}

func NewContext(db *sql.DB, store *sessions.CookieStore) *Context {
	return &Context{db, store}
}

func rootHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	session, err := c.store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if user, set := session.Values["user"]; set {
		fmt.Fprintf(w, "<h1>Welcome, %v</h1>", user)
	} else {
		renderTemplate(w, "login.html")
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	register(w, r, c, user, pass)
}

func loginHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	if login(w, r, c, user, pass) {
		fmt.Fprintf(w, "Successfully logged in!")
	} else {
		fmt.Fprintf(w, "Login failed.")
	}
}

func makeHandler(c *Context, fn func(http.ResponseWriter, *http.Request, *Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, c)
	}
}

func main() {
	os.Remove("./foo.db")

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c := NewContext(db, sessions.NewCookieStore([]byte("something-very-secret")))

	_, err = c.db.Query("create table users(username text primary key, password text)")
	if err != nil {
		log.Fatal(err)
		return
	}

	http.HandleFunc("/", makeHandler(c, rootHandler))
	http.HandleFunc("/register", makeHandler(c, registerHandler))
	http.HandleFunc("/login", makeHandler(c, loginHandler))

	fmt.Println("Serving webserver...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func register(w http.ResponseWriter, r *http.Request, c *Context, user, pass string) bool {
	_, err := c.db.Exec("insert into users values(?, ?)", user, pass)
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

func login(w http.ResponseWriter, r *http.Request, c *Context, user, pass string) bool {
	rows, err := c.db.Query("select * from users where username=?", user)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}
	for rows.Next() {
		var realUser, realPass string
		rows.Scan(&realUser, &realPass)

		if realUser == user && realPass == pass {
			// Get a session. We're ignoring the error resulted from decoding an
			// existing session: Get() always returns a session, even if empty.
			session, _ := c.store.Get(r, "session-name")
			// Set some session values.
			session.Values["user"] = user
			// Save it.
			session.Save(r, w)

			return true
		}
	}

	return false
}
