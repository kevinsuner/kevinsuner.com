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
	"strings"

	"github.com/yuin/goldmark"
)

func ViewArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/article/")
	if len(slug) < 0 {
		http.Error(w, fmt.Sprintf("failed to trim slug: %v", emptyString), http.StatusBadRequest)
		return
	}

	var article Article
	err := db.QueryRow(
		`SELECT created_at, updated_at, title, description, author, status, content FROM public.articles WHERE slug = $1`, slug).Scan(
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
		filepath.Join("views", "layouts", "header.tmpl"),
		filepath.Join("views", "layouts", "navbar.tmpl"),
		filepath.Join("views", "layouts", "footer.tmpl"),
		filepath.Join("views", "articles", "article.tmpl"),
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
	templateData := TemplateData{
		Meta: Meta{
			Description: article.Description,
			Author: article.Author,
			Type: "article",
			URL: fmt.Sprintf("https://%s", r.Host),
			Title: fmt.Sprintf("%s | Kevin Suñer", article.Title),
			CreatedAt: article.CreatedAt.String,
			UpdatedAt: article.UpdatedAt.String,},
		Article: article,
		HTML: template.HTML(html)}

	if err = t.Execute(&buf, templateData); err != nil {
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
		`SELECT id, title, slug, description, author, status, content FROM public.articles WHERE id = $1`, id).Scan(
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
		filepath.Join("views", "layouts", "admin_header.tmpl"),
		filepath.Join("views", "layouts", "navbar.tmpl"),
		filepath.Join("views", "layouts", "footer.tmpl"),
		filepath.Join("views", "articles", "edit.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, article); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func CreateArticle(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("create").ParseFiles(
		filepath.Join("views", "layouts", "admin_header.tmpl"),
		filepath.Join("views", "layouts", "navbar.tmpl"),
		filepath.Join("views", "layouts", "footer.tmpl"),
		filepath.Join("views", "articles", "create.tmpl"),
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

func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("admin_token")
	if errors.Is(err, http.ErrNoCookie) {
		t, err := template.New("login").ParseFiles(
			filepath.Join("views", "layouts", "admin_header.tmpl"),
			filepath.Join("views", "layouts", "navbar.tmpl"),
			filepath.Join("views", "layouts", "footer.tmpl"),
			filepath.Join("views", "admin", "login.tmpl"),
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
			filepath.Join("views", "layouts", "admin_header.tmpl"),
			filepath.Join("views", "layouts", "navbar.tmpl"),
			filepath.Join("views", "layouts", "footer.tmpl"),
			filepath.Join("views", "admin", "login.tmpl"),
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

	var pages int
	err = db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(id) / %d as pages FROM public.articles`, ARTICLES_LIMIT)).Scan(&pages)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to scan value: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("dashboard").Funcs(templateFuncs).ParseFiles(
		filepath.Join("views", "layouts", "admin_header.tmpl"),
		filepath.Join("views", "layouts", "navbar.tmpl"),
		filepath.Join("views", "layouts", "footer.tmpl"),
		filepath.Join("views", "admin", "dashboard.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	templateData := TemplateData{Pages: pages}

	if err = t.Execute(&buf, templateData); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}


func ProjectsPage(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT title, content, link, image, caption FROM public.projects ORDER BY id ASC`)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get projects: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		if err = rows.Scan(
			&project.Title,
			&project.Content,
			&project.Link,
			&project.Image,
			&project.Caption); err != nil {
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

	t, err := template.New("projects").ParseFiles(
		filepath.Join("views", "layouts", "header.tmpl"),
		filepath.Join("views", "layouts", "navbar.tmpl"),
		filepath.Join("views", "layouts", "footer.tmpl"),
		filepath.Join("views", "projects.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	templateData := TemplateData{
		Meta: Meta{
			Description: "This is a list of personal programming-related projects I've worked on my free time. Most of them are still being maintained, but some get more attention than others as my interests vary over time.",
			Author: "Kevin Suñer",
			Type: "website",
			URL: fmt.Sprintf("https://%s", r.Host),
			Title: "Projects | Kevin Suñer"},
		Projects: projects}

	if err = t.Execute(&buf, templateData); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func AboutPage(w http.ResponseWriter, r *http.Request) {
	var content string
	err := db.QueryRow(
		`SELECT content FROM public.pages WHERE title = 'about'`).Scan(&content)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get page: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("about").ParseFiles(
		filepath.Join("views", "layouts", "header.tmpl"),
		filepath.Join("views", "layouts", "navbar.tmpl"),
		filepath.Join("views", "layouts", "footer.tmpl"),
		filepath.Join("views", "about.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = goldmark.Convert([]byte(content), &buf); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse markdown: %v", err), http.StatusInternalServerError)
		return
	}
	html := buf.String()

	buf = bytes.Buffer{}
	templateData := TemplateData{
		Meta: Meta{
			Description: "Kevin is a software engineer working on stuff such as distributed systems, identity management and developer experience. He has been programming since the late 2000s after failing to cheat on a videogame called Lineage II.",
			Author: "Kevin Suñer",
			Type: "website",
			URL: fmt.Sprintf("https://%s", r.Host),
			Title: "About | Kevin Suñer"},
		HTML: template.HTML(html)}

	if err = t.Execute(&buf, templateData); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	var pages int
	err := db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(id) / %d as pages FROM public.articles WHERE status = 'published'`, ARTICLES_LIMIT)).Scan(&pages)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to scan value: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("homepage").Funcs(templateFuncs).ParseFiles(
		filepath.Join("views", "layouts", "header.tmpl"),
		filepath.Join("views", "layouts", "navbar.tmpl"),
		filepath.Join("views", "layouts", "footer.tmpl"),
		filepath.Join("views", "homepage.tmpl"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	templateData := TemplateData{
		Meta: Meta{
			Description: "Kevin is a software engineer working on stuff such as distributed systems, identity management and developer experience. He has been programming since the late 2000s after failing to cheat on a videogame called Lineage II.",
			Author: "Kevin Suñer",
			Type: "website",
			URL: fmt.Sprintf("https://%s", r.Host),
			Title: "Home | Kevin Suñer"},
		Pages: pages}

	if err = t.Execute(&buf, templateData); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func InitViews(mux *http.ServeMux) {
	/*** Private **/
	mux.HandleFunc(fmt.Sprintf("/%s", os.Getenv("ADMIN_URL")), AdminDashboard)
	mux.Handle("/create/article", CheckCookie(http.HandlerFunc(CreateArticle)))
	mux.Handle("/edit/article", CheckCookie(http.HandlerFunc(EditArticle)))

	/*** Public ***/
	mux.HandleFunc("/", HomePage)
	mux.HandleFunc("/about", AboutPage)
	mux.HandleFunc("/projects", ProjectsPage)
	mux.HandleFunc("/article/", ViewArticle)
}
