package main

import (
		"database/sql"
		"net/http"
		"html/template"
		"fmt"
	_	"github.com/mattn/go-sqlite3"
)


type User struct {
	ID int `json:"id"`
	Username string `json:"username"`
	Password string `json: "password"`
}



func createTable(qry string) {
		db, err := sql.Open("sqlite3", "wonderland.db")
		if err != nil {
			panic(err)
		}
		db.Exec("CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL UNIQUE, password TEXT NOT NULL);")
		db.Exec(qry)
		db.Close()

}


type MyH struct{}

func (h *MyH) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "test")
}

func handlefunc(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(w, nil)
}

func handlelog(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1> welcome !</h1>")
}

func auth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		if r.FormValue("username") == "" || r.FormValue("password") == "" {
			fmt.Fprintf(w, "<h1>error empty string!</h1>")
			return;
		}
		var usr User
		db, _ := sql.Open("sqlite3", "wonderland.db")
		smt, err := db.Prepare("SELECT username, password FROM users WHERE username = ?")
		if err != nil {
			panic(err)
		}
		defer smt.Close()
		defer db.Close()
		row := smt.QueryRow(r.FormValue("username"))
		err = row.Scan(&usr.Username, &usr.Password)

		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Fprint(w, "user cannot be found!")
				return
			}
			panic(err.Error())
		}
		if usr.Password != r.FormValue("password") {
				fmt.Fprint(w, "wrong password!")
				return
		}
		println(usr.Username, usr.Password)
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
		smt, _ := db.Prepare("SELECT COUNT(*) FROM users WHERE username=?")
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
		tmp, _ := smt.Exec(r.FormValue("username"),r.FormValue("password"))
		smt.Close()
		db.Close()
		temp, _ := tmp.LastInsertId()
		fmt.Fprint(w, temp)
		return
}

func main () {
//	test := MyH {}
	http.HandleFunc("/", handlefunc)
	http.Handle("/log", auth(http.HandlerFunc(handlelog)))
	http.HandleFunc("/register", registerhandler)
	http.HandleFunc("/api/register", registerapi)
	http.ListenAndServe(":8080", nil)
}
