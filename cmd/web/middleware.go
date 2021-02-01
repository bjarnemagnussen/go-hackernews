package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/justinas/nosurf"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			var redirect string
			switch r.Method {
			case "POST":
				// Add "goto" information to allow the login to return to the page it
				// came from and populate post data from "text".
				data := r.PostForm
				// app.session.Put(r, "goto", data.Get("goto"))
				app.session.Put(r, "text", data.Get("text"))
				redirect = data.Get("goto")

			case "GET":
				// app.session.Put(r, "goto", r.URL.Query().Get("goto"))
				redirect = r.URL.Query().Get("goto")
			}

			http.Redirect(w, r, "/user/login?goto="+redirect, 302)
			// http.Redirect(w, r, "/user/login?goto="+app.session.PopString(r, "goto"), 302)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.session.Exists(r, "authenticatedUserID") {
			next.ServeHTTP(w, r)
			return
		}

		user, err := app.user.GetFull(app.session.GetInt(r, "authenticatedUserID"))
		if err == models.ErrNoRecord || !user.Active {
			app.session.Remove(r, "authenticatedUserID")
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), contextAuthenticatedUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
