package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/yuin/goldmark"
)

var (
	errEmptyString			error = errors.New("empty string")
	errInvalidCredentials	error = errors.New("invalid username or password")
)

func GetPage(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query().Get("title")) <= 0 {
		http.Error(w, fmt.Sprintf("failed to parse title: %v", errEmptyString), http.StatusBadRequest)
		return
	}

	var content string
	err := db.QueryRow(
		`select content from pages where title = $1`, r.URL.Query().Get("title")).Scan(&content)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, fmt.Sprintf("failed to get page: %v", err), http.StatusInternalServerError)
		return
	}

	if len(content) > 0 {
		var buf bytes.Buffer
		if err = goldmark.Convert([]byte(content), &buf); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse markdown: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(buf.Bytes())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`
	<h2 class="color-blue-primary font-monoid-bold">¡Oopsie Daisy!</h2>
	<h5 class="font-monospace">Looks like there is no information to show yet :(</h5>`))
}

func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`delete from "articles" where id = $1`, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete article: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("/%s", os.Getenv("ADMIN_URL")))
}

func PutArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	var (
		title			string = r.Form.Get("title")
		slug			string = r.Form.Get("slug")
		description		string = r.Form.Get("description")
		author			string = r.Form.Get("author")
		status			string = r.Form.Get("status")
		content			string = r.Form.Get("content")
	)

	if err = checkEmptyString(title, slug, description, author, status, content); err != nil {
		http.Error(w, fmt.Sprintf("failed to validate form values: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(
		`update "articles" set updated_at=$1, title=$2, slug=$3, description=$4, author=$5, status=$6, content=$7 where id = $8`,
		time.Now().Format(time.RFC3339), title, slug, description, author, status, content, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update article: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`
	<div class="alert alert-primary" role="alert">
		<p>¡Hey! The article has been successfully edited</p>
		<hr>
		<a href="/%s" class="color-blue-primary mb-0">Back to Dashboard &#x2192;</a>
	</div>`, os.Getenv("ADMIN_URL"))))
}

func PostArticle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	var (
		title			string = r.Form.Get("title")
		slug			string = r.Form.Get("slug")
		description		string = r.Form.Get("description")
		author			string = r.Form.Get("author")
		status			string = r.Form.Get("status")
		content			string = r.Form.Get("content")
	)

	if err := checkEmptyString(title, slug, description, author, status, content); err != nil {
		http.Error(w, fmt.Sprintf("failed to validate form values: %v", err), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		`insert into "articles" (created_at, title, slug, description, author, status, content) values ($1, $2, $3, $4, $5, $6, $7)`,
		time.Now().Format(time.RFC3339), title, slug, description, author, status, content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to post article: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`
	<div class="alert alert-primary" role="alert">
		<p>¡Hooray! A new article has been created</p>
		<hr>
		<a href="/%s" class="color-blue-primary mb-0">Back to Dashboard &#x2192;</a>
	</div>`, os.Getenv("ADMIN_URL"))))
}

func GetArticles(w http.ResponseWriter, r *http.Request) {
	var err error
	var offset int = 0
	if len(r.URL.Query().Get("offset")) > 0 {
		offset, err = strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse int: %v", err), http.StatusBadRequest)
			return
		}
	}

	var (
		query string = fmt.Sprintf(`
			select id, created_at, updated_at, title, slug, description, author, status 
			from "articles"
			where status = 'published'
			order by created_at desc 
			limit %d offset %d`, ARTICLES_LIMIT, offset) 
		isAdmin bool
	)

	if len(r.URL.Query().Get("admin")) > 0 {
		val, err := strconv.ParseBool(r.URL.Query().Get("admin"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse bool: %v", err), http.StatusBadRequest)
			return
		}

		if val {
			cookie, err := r.Cookie("admin_token")
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(w, fmt.Sprintf("failed to authenticate: %v", err), http.StatusUnauthorized)
				return
			} else if err != nil {
				http.Error(w, fmt.Sprintf("failed to get cookie: %v", err), http.StatusInternalServerError)
				return
			}

			if cookie.Value != os.Getenv("ADMIN_TOKEN") {
				http.Error(w, fmt.Sprintf("failed to authenticate: %v", errInvalidToken), http.StatusUnauthorized)
				return
			}

			query = fmt.Sprintf(`
				select id, created_at, updated_at, title, slug, description, author, status 
				from "articles"
				order by created_at desc
				limit %d offset %d`, ARTICLES_LIMIT, offset)
			isAdmin = true
		}
	}

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get articles: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var article Article
		if err = rows.Scan(
			&article.ID,
			&article.CreatedAt,
			&article.UpdatedAt,
			&article.Title,
			&article.Slug,
			&article.Description,
			&article.Author,
			&article.Status); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan value: %v", err), http.StatusInternalServerError)
			return
		}
		articles = append(articles, article)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("failed while iterating: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("articles").ParseFiles(filepath.Join("views", "articles", "articles.tmpl"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"articles": articles,
		"is_admin": isAdmin,
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	if r.Form.Get("username") != os.Getenv("ADMIN_USER") ||
		r.Form.Get("password") != os.Getenv("ADMIN_PASS") {
		http.Error(w, fmt.Sprintf("failed to authenticate: %v", errInvalidCredentials), http.StatusUnauthorized)
		return
	}

	cookie := http.Cookie{}
	cookie.Name = "admin_token"
	cookie.Value = os.Getenv("ADMIN_TOKEN")
	cookie.Expires = time.Now().Add(time.Hour * 1)
	cookie.Secure = true
	cookie.HttpOnly = true
	cookie.Path = "/"
	http.SetCookie(w, &cookie)
	w.Header().Add("HX-Redirect", fmt.Sprintf("/%s", os.Getenv("ADMIN_URL")))
}

func InitEndpoints(mux *http.ServeMux) {
	/*** Private **/
	mux.Handle("/post/article", CheckCookie(http.HandlerFunc(PostArticle)))
	mux.Handle("/put/article", CheckCookie(http.HandlerFunc(PutArticle)))
	mux.Handle("/delete/article", CheckCookie(http.HandlerFunc(DeleteArticle)))

	/*** Public ***/
	mux.HandleFunc("/authenticate", Authenticate)
	mux.HandleFunc("/get/articles", GetArticles)
	mux.HandleFunc("/get/page", GetPage)
}
