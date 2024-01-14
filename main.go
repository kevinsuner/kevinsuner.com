package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const ARTICLES_LIMIT int = 5

var db *sql.DB
var templateFuncs = template.FuncMap{
	"Iterate": func(count int) []int {
		var numbers []int
		for i := 0; i < count; i++ {
			numbers = append(numbers, i)
		}
		return numbers
	},
	"Offset": func(num int) int {
		return num * ARTICLES_LIMIT 
	}}

func checkEmptyString(str ...string) error {
	for _, s := range str {
		if len(s) == 0 {
			return errEmptyString
		}
	}
	return nil
}

func main() {
	err := checkEmptyString(
		os.Getenv("PORT"),
		os.Getenv("ADMIN_URL"),
		os.Getenv("ADMIN_TOKEN"),
		os.Getenv("ADMIN_USER"),
		os.Getenv("ADMIN_PASS"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASS"),
		os.Getenv("PG_HOST"),
		os.Getenv("PG_NAME"))
	if err != nil {
		fmt.Fprintln(os.Stdout, errEmptyString.Error())
		os.Exit(1)
	}

	if os.Getenv("ENV") == "dev" {
		err = godotenv.Load()
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
			os.Exit(1)
		}
	}

	db, err = sql.Open("postgres", fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASS"),
		os.Getenv("PG_HOST"),
		os.Getenv("PG_NAME")))
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	
	InitEndpoints(mux)
	InitViews(mux)
	
	http.ListenAndServe(":"+os.Getenv("PORT"), mux)
}
