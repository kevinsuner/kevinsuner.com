package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
)

/*** Projects ***/

func CreateProject(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("create").ParseFiles(
		filepath.Join("templates", "layouts", "admin_header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "projects", "create.tmpl"),
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, nil); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func EditProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	var project Project
	err = db.QueryRow(
		`select id, title, link, image, caption, content from "projects" where id = $1`, id).Scan(
		&project.ID,
		&project.Title,
		&project.Link,
		&project.Image,
		&project.Caption,
		&project.Content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get project: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("edit").ParseFiles(
		filepath.Join("templates", "layouts", "admin_header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "projects", "edit.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"project": project,
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

/*** Pages ***/

func CreatePage(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("create").ParseFiles(
		filepath.Join("templates", "layouts", "admin_header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "pages", "create.tmpl"),
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, nil); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func EditPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	var page Page
	err = db.QueryRow(
		`select id, title, content from "pages" where id = $1`, id).Scan(
		&page.ID,
		&page.Title,
		&page.Content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get page: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("edit").ParseFiles(
		filepath.Join("templates", "layouts", "admin_header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "pages", "edit.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"page": page,
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

/*** Articles ***/

func CreateArticle(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("create").ParseFiles(
		filepath.Join("templates", "layouts", "admin_header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "articles", "create.tmpl"),
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, nil); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func ViewArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/article/")
	if len(slug) < 0 {
		http.Error(w, fmt.Sprintf("failed to trim slug: %v", errEmptyString), http.StatusBadRequest)
		return
	}

	var article Article
	err := db.QueryRow(
		`select created_at, updated_at, title, description, author, status, content from articles where slug = $1`, slug).Scan(
		&article.CreatedAt,
		&article.UpdatedAt,
		&article.Title,
		&article.Description,
		&article.Author,
		&article.Status,
		&article.Content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get article: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("article").ParseFiles(
		filepath.Join("templates", "layouts", "header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "articles", "article.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = goldmark.Convert([]byte(article.Content), &buf); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse markdown: %v", err), http.StatusInternalServerError)
		return
	}
	html := buf.String() 

	buf = bytes.Buffer{}
	if err = t.Execute(&buf, map[string]interface{}{
		"meta": Meta{
			Description: article.Description,
			Author: article.Author,
			Type: "article",
			URL: "https://"+r.Host,
			Title: fmt.Sprintf("%s | Kevin Suñer", article.Title),
			CreatedAt: article.CreatedAt.String,
			UpdatedAt: article.UpdatedAt.String,
		},
		"article": article,
		"html": template.HTML(html),
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func EditArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	var article Article
	err = db.QueryRow(
		`select id, title, slug, description, author, status, content from articles where id = $1`, id).Scan(
		&article.ID,
		&article.Title,
		&article.Slug,
		&article.Description,
		&article.Author,
		&article.Status,
		&article.Content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get article: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("edit").ParseFiles(
		filepath.Join("templates", "layouts", "admin_header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "articles", "edit.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"article": article,
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

/*** Admin ***/

func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("admin_token")
	if errors.Is(err, http.ErrNoCookie) {
		t, err := template.New("login").ParseFiles(
			filepath.Join("templates", "layouts", "admin_header.tmpl"),
			filepath.Join("templates", "layouts", "navbar.tmpl"),
			filepath.Join("templates", "layouts", "footer.tmpl"),
			filepath.Join("templates", "admin", "login.tmpl"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		if err = t.Execute(&buf, nil); err != nil {
			http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(buf.Bytes())
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("failed to get cookie: %v", err), http.StatusInternalServerError)
		return
	}

	if cookie.Value != os.Getenv("ADMIN_TOKEN") {
		t, err := template.New("login").ParseFiles(
			filepath.Join("templates", "layouts", "admin_header.tmpl"),
			filepath.Join("templates", "layouts", "navbar.tmpl"),
			filepath.Join("templates", "layouts", "footer.tmpl"),
			filepath.Join("templates", "admin", "login.tmpl"),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		if err = t.Execute(&buf, nil); err != nil {
			http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(buf.Bytes())
		return
	}

	var pages float64
	err = db.QueryRow(fmt.Sprintf(`
		select count(id)::float / %d as pages from articles`, ARTICLES_LIMIT)).Scan(&pages)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to scan value: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("dashboard").Funcs(templateFuncs).ParseFiles(
		filepath.Join("templates", "layouts", "admin_header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "admin", "dashboard.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"pages": int(math.Ceil(pages)),
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

/*** Views ***/

func Projects(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("projects").ParseFiles(
		filepath.Join("templates", "layouts", "header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "projects.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"meta": Meta{
			Description: "This is a list of personal programming-related projects I've worked on my free time. Most of them are still being maintained, but some get more attention than others as my interests vary over time.",
			Author: "Kevin Suñer",
			Type: "website",
			URL: "https://"+r.Host,
			Title: "Projects | Kevin Suñer",
		},
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func About(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("about").ParseFiles(
		filepath.Join("templates", "layouts", "header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "about.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"meta": Meta{
			Description: "Kevin is a software engineer working on stuff such as distributed systems, identity management and developer experience. He has been programming since the late 2000s after failing to cheat on a videogame called Lineage II.",
			Author: "Kevin Suñer",
			Type: "website",
			URL: "https://"+r.Host,
			Title: "About | Kevin Suñer",
		},
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func Home(w http.ResponseWriter, r *http.Request) {
	var pages float64
	err := db.QueryRow(fmt.Sprintf(`
		select count(id)::float / %d as pages from articles where status = 'published'`, ARTICLES_LIMIT)).Scan(&pages)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to scan value: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("homepage").Funcs(templateFuncs).ParseFiles(
		filepath.Join("templates", "layouts", "header.tmpl"),
		filepath.Join("templates", "layouts", "navbar.tmpl"),
		filepath.Join("templates", "layouts", "footer.tmpl"),
		filepath.Join("templates", "homepage.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"meta": Meta{
			Description: "Kevin is a software engineer working on stuff such as distributed systems, identity management and developer experience. He has been programming since the late 2000s after failing to cheat on a videogame called Lineage II.",
			Author: "Kevin Suñer",
			Type: "website",
			URL: "https://"+r.Host,
			Title: "Home | Kevin Suñer",
		},
		"pages": int(math.Ceil(pages)),
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func InitViews(mux *http.ServeMux) {
	/*** Projects ***/
	mux.Handle("/create/project", CheckCookie(http.HandlerFunc(CreateProject)))
	mux.Handle("/edit/project", CheckCookie(http.HandlerFunc(EditProject)))
	
	/*** Pages ***/
	mux.Handle("/create/page", CheckCookie(http.HandlerFunc(CreatePage)))
	mux.Handle("/edit/page", CheckCookie(http.HandlerFunc(EditPage)))

	/*** Articles ***/
	mux.Handle("/create/article", CheckCookie(http.HandlerFunc(CreateArticle)))
	mux.HandleFunc("/article/", ViewArticle)
	mux.Handle("/edit/article", CheckCookie(http.HandlerFunc(EditArticle)))

	/*** Admin ***/
	mux.HandleFunc(fmt.Sprintf("/%s", os.Getenv("ADMIN_URL")), AdminDashboard)

	/*** Views ***/
	mux.HandleFunc("/projects", Projects)
	mux.HandleFunc("/about", About)
	mux.HandleFunc("/", Home)
}
