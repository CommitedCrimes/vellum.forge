package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/home.jet")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) blogIndex(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Load blog posts from content directory (with app.config.dataDir as base directory)
	blogPosts, err := app.contentLoader.LoadBlogPosts(app.config.dataDir)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data["BlogPosts"] = blogPosts

	err = app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/blog/index.jet")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) blogPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Load blog post by slug
	blogPost, err := app.contentLoader.LoadBlogPost(app.config.dataDir, slug)
	if err != nil {
		app.notFound(w, r)
		return
	}

	data := app.newTemplateData(r)
	data["Post"] = blogPost

	err = app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/blog/post.jet")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) page(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Load page by slug
	page, err := app.contentLoader.LoadPage(app.config.dataDir, slug)
	if err != nil {
		app.notFound(w, r)
		return
	}

	data := app.newTemplateData(r)
	data["Page"] = page

	err = app.jetRenderer.RenderPage(w, http.StatusOK, data, "pages/page.jet")
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
