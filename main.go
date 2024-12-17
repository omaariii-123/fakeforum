package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json: "password"`
}
type Session struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	UserID   int    `json:"userid"`
	// ExpDate	time.Time	`json:"expdate"`
}

func createTable(qry string) {
	db, err := sql.Open("sqlite3", "wonderland.db")
	if err != nil {
		println(err.Error())
	}
	db.Exec(qry)
	db.Close()
}

type MyH struct{}

func (h *MyH) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "test")
}

func handlefunc(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		println(cookie.Name, cookie.Value)
		http.Redirect(w, r, "/dash", http.StatusFound)
	}
	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(w, nil)
}

func handlelog(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dash", http.StatusFound)
}

func createSession(w http.ResponseWriter, user User) string {
	id, _ := uuid.NewV7()
	db, _ := sql.Open("sqlite3", "wonderland.db")
	smt, _ := db.Prepare("INSERT INTO sessions(id, username, userId) VALUES(?,?,?)")
	println("**********", id.String(), user.Username, user.ID)
	_, err := smt.Exec(id.String(), user.Username, user.ID)
	if err != nil {
		println("**********-------------------", err.Error())
	}
	smt.Close()
	db.Close()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    id.String(),
		Path:     "/",
		HttpOnly: true,
	})
	return id.String()
}

func auth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("username") == "" || r.FormValue("password") == "" {
			fmt.Fprintf(w, "<h1>error empty creds!</h1>")
			return
		}
		var usr User
		db, _ := sql.Open("sqlite3", "wonderland.db")
		smt, err := db.Prepare("SELECT id, username, password FROM users WHERE username = ?")
		if err != nil {
			panic(err)
		}
		row := smt.QueryRow(r.FormValue("username"))
		err = row.Scan(&usr.ID, &usr.Username, &usr.Password)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Fprint(w, "user cannot be found!")
				return
			}
			panic(err.Error())
		}
		smt.Close()
		if usr.Password != r.FormValue("password") {
			fmt.Fprint(w, "wrong password!")
			return
		}
		var tmp string
		var name string
		tmpc := ""
		println(usr.Username, usr.Password)
		cookie, err := r.Cookie("session_id")
		if err == nil {
			tmpc = cookie.Value
		}
		smt, _ = db.Prepare("SELECT id, username FROM sessions WHERE id = ?")
		row = smt.QueryRow(tmpc)
		err = row.Scan(&tmp, &name)
		if err == sql.ErrNoRows {
			fmt.Println("created session's id ", createSession(w, usr))
		}
		println(name)
		smt.Close()
		db.Close()
		handler.ServeHTTP(w, r)
	})
}

func registerhandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("register.html")
	tmpl.Execute(w, nil)
}

func check_exist(username string) bool {
	var name string
	db, _ := sql.Open("sqlite3", "wonderland.db")
	smt, _ := db.Prepare("SELECT username FROM users WHERE username=?")
	defer smt.Close()
	defer db.Close()
	row := smt.QueryRow(username)
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		return false
	}
	return true
}

func registerapi(w http.ResponseWriter, r *http.Request) {
	if check_exist(r.FormValue("username")) {
		fmt.Fprint(w, "user already exists!")
		return
	}
	test := "INSERT INTO users(username, password) VALUES(?,?);"
	db, _ := sql.Open("sqlite3", "wonderland.db")
	smt, _ := db.Prepare(test)
	tmp, _ := smt.Exec(r.FormValue("username"), r.FormValue("password"))
	smt.Close()
	db.Close()
	temp, _ := tmp.LastInsertId()
	fmt.Fprint(w, temp)
	return
}

func handling(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	var name string
	var tmp string
	if err == nil {
		db, _ := sql.Open("sqlite3", "wonderland.db")
		smt, _ := db.Prepare("SELECT id, username FROM sessions WHERE id = ?")
		row := smt.QueryRow(cookie.Value)
		row.Scan(&tmp, &name)
		smt.Close()
		db.Close()
	}
	tmpl, _ := template.ParseFiles("dash.html")
	tmpl.Execute(w, name)
}

func deleteSession(r *http.Request) {
	cookie, _ := r.Cookie("session_id")
	db, _ := sql.Open("sqlite3", "wonderland.db")
	smt, _ := db.Prepare("DELETE FROM sessions WHERE id = ?")
	smt.Exec(cookie.Value)
	smt.Close()
	db.Close()
}

func handlelogout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	deleteSession(r)
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	//	test := MyH {}
	createTable("CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL UNIQUE, password TEXT NOT NULL);")
	createTable("CREATE TABLE IF NOT EXISTS sessions(id TEXT PRIMARY KEY, username TEXT, userId INTEGER)")
	http.HandleFunc("/", handlefunc)
	http.HandleFunc("/dash", handling)
	http.Handle("/log", auth(http.HandlerFunc(handlelog)))
	http.HandleFunc("/logout", handlelogout)

	http.HandleFunc("/register", registerhandler)
	http.HandleFunc("/api/register", registerapi)
	http.ListenAndServe(":8080", nil)
}
