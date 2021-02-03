package main

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

type router struct {
	*httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
	r.HEAD(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
	r.POST(path, wrapHandler(handler))
}

// newRouter is a wrapper for a httprouter.Router to seamlessly use
// http.Handlers when defining routes.
func newRouter() *router {
	return &router{httprouter.New()}
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), contextParams, ps)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)
	// dynamicMiddleware := alice.New(app.session.Enable, app.authenticate)

	router := newRouter()
	router.Get("/", dynamicMiddleware.ThenFunc(app.home))
	router.Get("/news/:page", dynamicMiddleware.ThenFunc(app.home))
	router.Get("/newest", dynamicMiddleware.ThenFunc(app.showNewest))
	router.Get("/newest/:next/:page", dynamicMiddleware.ThenFunc(app.showNewest))
	router.Get("/comments", dynamicMiddleware.ThenFunc(app.showNewestComments))
	router.Get("/comments/:next/:page", dynamicMiddleware.ThenFunc(app.showNewestComments))

	router.Get("/submissions/:foruser", dynamicMiddleware.ThenFunc(app.showNewest))
	router.Get("/submissions/:foruser/:next/:page", dynamicMiddleware.ThenFunc(app.showNewest))
	router.Get("/threads/:foruser", dynamicMiddleware.ThenFunc(app.showUserComments))
	router.Get("/threads/:foruser/:next/:page", dynamicMiddleware.ThenFunc(app.showUserComments))
	router.Get("/from/:domain", dynamicMiddleware.ThenFunc(app.showFromDomain))
	router.Get("/from/:domain/:next/:page", dynamicMiddleware.ThenFunc(app.showFromDomain))
	router.Get("/ask/:order", dynamicMiddleware.ThenFunc(app.showAskUs))
	router.Get("/ask/:order/:next/:page", dynamicMiddleware.ThenFunc(app.showAskUs))
	router.Get("/show/:order", dynamicMiddleware.ThenFunc(app.showShowUs))
	router.Get("/show/:order/:next/:page", dynamicMiddleware.ThenFunc(app.showShowUs))

	router.Get("/items/:postid", dynamicMiddleware.ThenFunc(app.showPost))
	router.Post("/items/:postid/upvote", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.upvotePost))
	router.Get("/items/:postid/comments/:commentid", dynamicMiddleware.ThenFunc(app.createReplyForm))
	router.Post("/items/:postid/comments/:commentid", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createReply))
	router.Post("/items/:postid/comments/:commentid/upvote", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.upvoteComment))

	router.Get("/submit", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createPostForm))
	router.Post("/submit", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createPost))

	router.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	router.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	router.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	router.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	router.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser))

	router.Get("/user/profile/:username", dynamicMiddleware.ThenFunc(app.showProfile))
	router.Get("/user/edit", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.editProfileForm))
	router.Post("/user/edit", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.editProfile))

	router.ServeFiles("/static/*filepath", http.Dir("./ui/static/"))

	router.Get("/guidelines", dynamicMiddleware.ThenFunc(app.guidelines))
	router.Get("/acknowledgements", dynamicMiddleware.ThenFunc(app.acknowledgements))
	router.Get("/faq", dynamicMiddleware.ThenFunc(app.faq))
	router.Get("/show-rules", dynamicMiddleware.ThenFunc(app.showUsRules))

	return standardMiddleware.Then(router)
}
