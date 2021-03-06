package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/MninaTB/bookings/internal/config"
	"github.com/MninaTB/bookings/internal/models"
	"github.com/justinas/nosurf"
)

// this is needed, so we can pass our own functions into the templates
var functions = template.FuncMap{}

var app *config.AppConfig
var pathToTemplates = "./templates"

// NewTemplates sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

// AddDefaultData returns the default template data that we want to be able to pass to every template
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")

	td.CSRFToken = nosurf.Token(r)
	return td
}

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template

	// check if the developement mode is on (UseCache would be false in the
	// main function) if so, we want to rebuild the template cache, instead of
	// pulling the template out of the map (the one in the config file)
	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		// creates new template cache
		tc, _ = CreateTemplateCache()
	}

	// check if the pagename exists
	t, ok := tc[tmpl]
	if !ok {
		return errors.New("can't get template from cache")
	}

	buf := new(bytes.Buffer)

	// default data were passed to the template
	td = AddDefaultData(td, r)

	// the current used template gets stored in bytes
	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
		return err
	}

	return nil
}

// CreateTemplateCache - creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {

	// This Map holds the ready to use templates and is searchable through the pagenames
	myChache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.html", pathToTemplates))
	if err != nil {
		return myChache, err
	}

	for _, page := range pages {
		// this extracts the name of the page out of the full path
		name := filepath.Base(page)

		// this is to give the template functions that we want into the template
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myChache, err
		}

		// looking if this template matches any layouts
		// - so we are checking if we should use a layout to this template
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
		if err != nil {
			return myChache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.html", pathToTemplates))
			if err != nil {
				return myChache, err
			}
		}

		myChache[name] = ts
	}

	return myChache, nil
}
