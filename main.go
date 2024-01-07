package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const ARTICLES_LIMIT int = 3

var db *sql.DB

var templateFuncs = template.FuncMap{
	"Iterate": func(count int) []int {
		var numbers []int
		for i := 0; i <= count; i++ {
			numbers = append(numbers, i)
		}
		return numbers
	},
	"Offset": func(num int) int {
		return num * ARTICLES_LIMIT
	}}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("[ERROR] failed to load .env file: %v\n", err)
	}

	db, err = sql.Open("postgres", fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("PQ_USER"), os.Getenv("PQ_PASS"), os.Getenv("PQ_IP"), os.Getenv("PQ_NAME")))
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize db: %v\n", err)
	}

	log.Println("[INFO] successfully connected to db")
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	
	InitEndpoints(mux)
	InitViews(mux)
	
	log.Printf("[INFO] started http server at port %s\n", os.Getenv("PORT"))
	http.ListenAndServe(os.Getenv("PORT"), mux)
}
