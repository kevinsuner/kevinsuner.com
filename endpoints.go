package main

import (
	"bytes"
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

/*** Projects ***/

func PostProject(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	var (
		title	string = r.Form.Get("title")
		link	string = r.Form.Get("link")
		image	string = r.Form.Get("image")
		caption	string = r.Form.Get("caption")
		content	string = r.Form.Get("content")
	)

	if err := checkEmptyString(title, link, image, caption, content); err != nil {
		http.Error(w, fmt.Sprintf("failed to validate form values: %v", err), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		`insert into "projects" (created_at, title, link, image, caption, content) values ($1, $2, $3, $4, $5, $6)`,
		time.Now().Format(time.RFC3339), title, link, image, caption, content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to post project: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`
	<div class="alert alert-primary" role="alert">
		<p>¡Hooray! A new project has been created</p>
		<hr>
		<a href="/%s" class="color-blue-primary mb-0">Back to Dashboard &#x2192;</a>
	</div>`, os.Getenv("ADMIN_URL"))))
}

func GetProjects(w http.ResponseWriter, r *http.Request) {
	var isAdmin bool
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

			isAdmin = true
		}
	}

	rows, err := db.Query(`select id, title, link, image, caption, content from "projects" order by id desc`)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get projects: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		if err = rows.Scan(
			&project.ID,
			&project.Title,
			&project.Link,
			&project.Image,
			&project.Caption,
			&project.Content); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan value: %v", err), http.StatusInternalServerError)
			return
		}
	
		var buf bytes.Buffer
		if err = goldmark.Convert([]byte(project.Content), &buf); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse markdown: %v", err), http.StatusInternalServerError)
			return
		}

		project.HTML = template.HTML(buf.String())
		projects = append(projects, project)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("failed while iterating: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("projects").ParseFiles(filepath.Join("templates", "projects", "projects.tmpl"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"projects": projects,
		"is_admin": isAdmin,
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func PutProject(w http.ResponseWriter, r *http.Request) {
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
		title	string = r.Form.Get("title")
		link	string = r.Form.Get("link")
		image	string = r.Form.Get("image")
		caption	string = r.Form.Get("caption")
		content	string = r.Form.Get("content")
	)

	if err = checkEmptyString(title, link, image, caption, content); err != nil {
		http.Error(w, fmt.Sprintf("failed to validate form values: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(
		`update "projects" set updated_at=$1, title=$2, link=$3, image=$4, caption=$5, content=$6 where id = $7`,
		time.Now().Format(time.RFC3339), title, link, image, caption, content, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update project: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`
	<div class="alert alert-primary" role="alert">
		<p>¡Hey! The project has been successfully edited</p>
		<hr>
		<a href="/%s" class="color-blue-primary mb-0">Back to Dashboard &#x2192;</a>
	</div>`, os.Getenv("ADMIN_URL"))))
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`delete from "projects" where id = $1`, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete project: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("/%s", os.Getenv("ADMIN_URL")))
}

/*** Pages ***/

func PostPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	var (
		title	string = r.Form.Get("title")
		content	string = r.Form.Get("content")
	)

	if err := checkEmptyString(title, content); err != nil {
		http.Error(w, fmt.Sprintf("failed to validate form values: %v", err), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		`insert into "pages" (created_at, title, content) values ($1, $2, $3)`,
		time.Now().Format(time.RFC3339), title, content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to post page: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`
	<div class="alert alert-primary" role="alert">
		<p>¡Hooray! A new page has been created</p>
		<hr>
		<a href="/%s" class="color-blue-primary mb-0">Back to Dashboard &#x2192;</a>
	</div>`, os.Getenv("ADMIN_URL"))))
}

func GetPages(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`select id, title from "pages"`)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get pages: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pages []Page
	for rows.Next() {
		var page Page
		if err = rows.Scan(&page.ID, &page.Title); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan value: %v", err), http.StatusInternalServerError)
			return
		}
		pages = append(pages, page)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("failed while iterating: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("pages").ParseFiles(filepath.Join("templates", "pages", "pages.tmpl"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"pages": pages,
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func GetPage(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query().Get("title")) <= 0 {
		http.Error(w, fmt.Sprintf("failed to parse title: %v", errEmptyString), http.StatusBadRequest)
		return
	}

	var content string
	err := db.QueryRow(
		`select content from "pages" where title = $1`, r.URL.Query().Get("title")).Scan(&content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get page: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = goldmark.Convert([]byte(content), &buf); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse markdown: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func PutPage(w http.ResponseWriter, r *http.Request) {
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
		title	string = r.Form.Get("title")
		content	string = r.Form.Get("content")
	)

	if err = checkEmptyString(title, content); err != nil {
		http.Error(w, fmt.Sprintf("failed to validate form values: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(
		`update "pages" set updated_at=$1, title=$2, content=$3 where id = $4`,
		time.Now().Format(time.RFC3339), title, content, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update page: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`
	<div class="alert alert-primary" role="alert">
		<p>¡Hey! The page has been successfully edited</p>
		<hr>
		<a href="/%s" class="color-blue-primary mb-0">Back to Dashboard &#x2192;</a>
	</div>`, os.Getenv("ADMIN_URL"))))
}

func DeletePage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`delete from "pages" where id = $1`, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete page: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("/%s", os.Getenv("ADMIN_URL")))
}

/*** Articles ***/

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

	t, err := template.New("articles").ParseFiles(filepath.Join("templates", "articles", "articles.tmpl"))
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

/*** Admin ***/

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
	/*** Projects ***/
	mux.Handle("/post/project", CheckCookie(http.HandlerFunc(PostProject)))
	mux.HandleFunc("/get/projects", GetProjects)
	mux.Handle("/put/project", CheckCookie(http.HandlerFunc(PutProject)))
	mux.Handle("/delete/project", CheckCookie(http.HandlerFunc(DeleteProject)))

	/*** Pages ***/
	mux.Handle("/post/page", CheckCookie(http.HandlerFunc(PostPage)))
	mux.Handle("/get/pages", CheckCookie(http.HandlerFunc(GetPages)))
	mux.HandleFunc("/get/page", GetPage)
	mux.Handle("/put/page", CheckCookie(http.HandlerFunc(PutPage)))
	mux.Handle("/delete/page", CheckCookie(http.HandlerFunc(DeletePage)))

	/*** Articles ***/
	mux.Handle("/post/article", CheckCookie(http.HandlerFunc(PostArticle)))
	mux.HandleFunc("/get/articles", GetArticles)
	mux.Handle("/put/article", CheckCookie(http.HandlerFunc(PutArticle)))
	mux.Handle("/delete/article", CheckCookie(http.HandlerFunc(DeleteArticle)))

	/*** Admin **/
	mux.HandleFunc("/authenticate", Authenticate)
}
