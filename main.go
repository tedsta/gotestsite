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

type User struct {
	username string
	password string
	email    string
}

func NewUser(username, password, email string) *User {
	return &User{username, password, email}
}

func rootHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	session, err := c.store.Get(r, "gositetest-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if user, set := session.Values["user"]; set {
		fmt.Fprintf(w, "<h1>Welcome, %v</h1><br>", user)
		fmt.Fprintf(w, "<a href=\"/dologout\">Logout</a>")
	} else {
		renderTemplate(w, "index.html")
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	renderTemplate(w, "register.html")
}

func loginHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	renderTemplate(w, "login.html")
}

func doregisterHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	// Create a the user from the submitted form
	u := &User{r.FormValue("user"), r.FormValue("pass"), r.FormValue("email")}

	// Register the user
	if register(w, r, c, u) {
		fmt.Fprintf(w, "Successfully registered!")
	} else {
		fmt.Fprintf(w, "Couldn't register you.")
	}
}

func dologinHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	if login(w, r, c, user, pass) {
		fmt.Fprintf(w, "Successfully logged in!")
	} else {
		fmt.Fprintf(w, "Login failed.")
	}
}

func dologoutHandler(w http.ResponseWriter, r *http.Request, c *Context) {
	logout(w)
	fmt.Fprintf(w, "Successfully logged out!")
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

	_, err = c.db.Exec("create table users(username text primary key, password text, email text)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", makeHandler(c, rootHandler))
	http.HandleFunc("/register", makeHandler(c, registerHandler))
	http.HandleFunc("/login", makeHandler(c, loginHandler))
	http.HandleFunc("/doregister", makeHandler(c, doregisterHandler))
	http.HandleFunc("/dologin", makeHandler(c, dologinHandler))
	http.HandleFunc("/dologout", makeHandler(c, dologoutHandler))

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

func register(w http.ResponseWriter, r *http.Request, c *Context, u *User) bool {
	tx, err := c.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	_, err = tx.Exec("insert into users values(?, ?, ?)", u.username, u.password, u.email)
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

func login(w http.ResponseWriter, r *http.Request, c *Context, user, pass string) bool {
	rows, err := c.db.Query("select username, password from users where username=?", user)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var realUser, realPass string
		if err := rows.Scan(&realUser, &realPass); err != nil {
			log.Fatal(err)
		}

		if realUser == user && realPass == pass {
			session, _ := c.store.Get(r, "gositetest-session")
			session.Values["user"] = user
			session.Save(r, w)

			return true
		}
	}

	return false
}

func logout(w http.ResponseWriter) {
	// gorilla/sessions doesn't let us delete a session D: do it manually
	http.SetCookie(w, &http.Cookie{Name: "gositetest-session", MaxAge: -1, Path: "/"})
}
