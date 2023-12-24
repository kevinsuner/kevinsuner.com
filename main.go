package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*** data ***/

const TEMPLATES_DIR string = "templates"
const PARTIALS_DIR string = "partials"

var db *sql.DB

/*** endpoints  ***/

func test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<h1 id="welcome-msg">Hello from Golang and HTMX</h1>`))
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("index").ParseFiles(
		filepath.Join(TEMPLATES_DIR, PARTIALS_DIR, "header.html"),
		filepath.Join(TEMPLATES_DIR, PARTIALS_DIR, "footer.html"),
		filepath.Join(TEMPLATES_DIR, "index.html"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] failed to parse template")
		return
	}
	
	var buf bytes.Buffer
	err = t.Execute(&buf, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("[ERROR] failed to execute template")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

/*** init ***/

func init() {
	err := godotenv.Load()
	if err != nil { panic(err) }

	var name string = os.Getenv("PQ_NAME")
	var user string = os.Getenv("PQ_USER")
	var pass string = os.Getenv("PQ_PASS")
	
	db, err = sql.Open("postgres", fmt.Sprintf("postgresql://%s:%s@tcp/%s?sslmode=disable", user, pass, name))
	if err != nil { panic(err) }

	log.Println("[INFO] connected to db successfully")
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", index)
	mux.HandleFunc("/test", test)
	http.ListenAndServe(":8080", mux)
}
