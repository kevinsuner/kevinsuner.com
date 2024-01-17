package main

import (
	"database/sql"
	"html/template"
)

type Article struct {
	ID uint
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
	Title string
	Slug string
	Description string
	Author string
	Status string
	Content string
}

type Page struct {
	ID uint
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
	Title string
	Content string
}

type Project struct {
	ID uint
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
	Title string
	Link string
	Image string
	Caption string
	Content string
	HTML template.HTML
}

type Meta struct {
	Description string
	Author string
	Type string
	URL string
	Title string
	CreatedAt string
	UpdatedAt string
}

type TemplateData struct {
	Meta Meta
	Article Article
	Articles []Article
	Projects []Project
	HTML template.HTML
	IsAdmin bool
	Pages int
}
