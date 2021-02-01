package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/justinas/nosurf"
)

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the user.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
	}

	buf := new(bytes.Buffer)

	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	buf.WriteTo(w)
}

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}

	// TODO?
	// td.CurrentYear = time.Now().Year()

	td.CSRFToken = nosurf.Token(r)

	td.Flash = app.session.PopString(r, "flash")

	td.IsAuthenticated = app.isAuthenticated(r)

	td.User = app.authenticatedUser(r)

	td.CurrentURL = r.RequestURI[1:]

	td.SiteName = app.config.Template.SiteName
	td.SiteShort = app.config.Template.SiteShort
	td.Generator = app.config.Template.Generator
	td.ThemeColor = app.config.Template.ThemeColor
	td.FooterText = app.config.Template.FooterText

	return td
}

func (app *application) isAuthenticated(r *http.Request) bool {
	authenticatedUser, ok := r.Context().Value(contextAuthenticatedUser).(*models.User)
	if !ok || authenticatedUser.ID == 0 {
		return false
	}

	return true
}

func (app *application) authenticatedUser(r *http.Request) *models.User {
	authenticatedUser, ok := r.Context().Value(contextAuthenticatedUser).(*models.User)
	if !ok || authenticatedUser.ID == 0 {
		return &models.User{}
	}

	return authenticatedUser
}
