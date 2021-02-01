package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/forms"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
	"github.com/julienschmidt/httprouter"
)

// TODO: Make Show/Ask BH: handler have a points threshold
// TODO: admin/moderator accounts that can flag submissions
// TODO: power user accounts that can vote to flag submissions

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	page, err := strconv.Atoi(params.ByName("page"))
	if err != nil {
		page = 0
	} else if page < 0 || page >= 10 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	posts, err := app.posts.Popular(user.ID, page*30)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var moreURL string
	if len(posts) == 31 && page < 10 {
		moreURL = strconv.Itoa(page + 1)
		posts = posts[:30]
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Page:      "news",
		Posts:     posts,
		StartRank: page*30 + 1,
		MoreURL:   moreURL,
	})
}

func (app *application) showNewest(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	lastID, err := strconv.Atoi(params.ByName("next"))
	if err != nil || lastID < 1 {
		lastID = 0
	}

	pageNum, err := strconv.Atoi(params.ByName("page"))
	if err != nil || pageNum < 0 {
		pageNum = 0
	}

	forUserID, err := app.user.GetIDByUsername(params.ByName("foruser"))
	if err != nil {
		forUserID = 0
	}

	forUser, err := app.user.GetFull(forUserID)
	if err != nil {
		forUser = &models.User{ID: 0}
	}

	page := "newest"
	if forUserID != 0 {
		page = forUser.Username + "'s submissions"
	}

	posts, err := app.posts.Latest(user.ID, lastID, forUser.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var moreURL string
	if len(posts) == 31 {
		moreURL = strconv.Itoa(posts[len(posts)-1].ID) + "/" + strconv.Itoa(pageNum+1)
		posts = posts[:30]
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Page:      page,
		Title:     strings.Title(page),
		Posts:     posts,
		StartRank: pageNum*30 + 1,
		MoreURL:   moreURL,
	})
}

func (app *application) showFromDomain(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	lastID, err := strconv.Atoi(params.ByName("next"))
	if err != nil || lastID < 1 {
		lastID = 0
	}

	page, err := strconv.Atoi(params.ByName("page"))
	if err != nil || page < 0 {
		page = 0
	}

	posts, err := app.posts.GetFromDomain(user.ID, lastID, params.ByName("domain"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	var moreURL string
	if len(posts) == 31 {
		moreURL = strconv.Itoa(posts[len(posts)-1].ID) + "/" + strconv.Itoa(page+1)
		posts = posts[:30]
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Page:      "from",
		Title:     "Submissions From " + params.ByName("domain"),
		Posts:     posts,
		StartRank: page*30 + 1,
		MoreURL:   moreURL,
	})
}

func (app *application) showShowBH(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	page, err := strconv.Atoi(params.ByName("page"))
	if err != nil {
		page = 0
	} else if page < 0 || page >= 10 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	posts, err := app.posts.GetForType(user.ID, page*30, mysql.ShowPost)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var moreURL string
	if len(posts) == 31 && page < 10 {
		moreURL = strconv.Itoa(page + 1)
		posts = posts[:30]
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Page:      "show",
		Title:     "Show",
		Posts:     posts,
		StartRank: page*30 + 1,
		MoreURL:   moreURL,
	})
}

func (app *application) showAskBH(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	page, err := strconv.Atoi(params.ByName("page"))
	if err != nil {
		page = 0
	} else if page < 0 || page >= 10 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	posts, err := app.posts.GetForType(user.ID, page*30, mysql.AskPost)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var moreURL string
	if len(posts) == 31 && page < 10 {
		moreURL = strconv.Itoa(page + 1)
		posts = posts[:30]
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Page:      "ask",
		Title:     "Ask",
		Posts:     posts,
		StartRank: page*30 + 1,
		MoreURL:   moreURL,
	})
}

func (app *application) showPost(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	id, err := strconv.Atoi(params.ByName("postid"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	p, err := app.posts.GetByID(user.ID, id)
	if err == models.ErrNoRecord || p.Deleted == 1 {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	c, err := app.comments.ForPost(user.ID, id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	f := forms.New(make(url.Values))
	// Populate the comment-form with text if saved in the session (due to
	// redirect from login)
	f.Add("text", app.session.PopString(r, "text"))

	app.render(w, r, "show.page.tmpl", &templateData{
		Post:     p,
		Title:    strings.Title(p.Title),
		Comments: c,
		Form:     f,
	})
}

func (app *application) showNewestComments(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	lastID, err := strconv.Atoi(params.ByName("next"))
	if err != nil || lastID < 1 {
		lastID = 0
	}

	pageNum, err := strconv.Atoi(params.ByName("page"))
	if err != nil || pageNum < 0 {
		pageNum = 0
	}

	page := "comments"

	comments, err := app.comments.Latest(user.ID, lastID)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	var moreURL string
	if len(comments) == 30 {
		moreURL = strconv.Itoa(comments[len(comments)-1].ID) + "/" + strconv.Itoa(pageNum+1)
	}

	app.render(w, r, "comments.page.tmpl", &templateData{
		Page:      page,
		Title:     "New Comments",
		Comments:  comments,
		StartRank: pageNum*30 + 1,
		MoreURL:   moreURL,
	})
}

func (app *application) showUserComments(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	lastID, err := strconv.Atoi(params.ByName("next"))
	if err != nil || lastID < 1 {
		lastID = 0
	}
	//fmt.Println("lastID: ", lastID)

	// pageNum, err := strconv.Atoi(params.ByName("page"))
	// if err != nil || pageNum < 0 {
	// 	pageNum = 0
	// }

	forUserID, err := app.user.GetIDByUsername(params.ByName("foruser"))
	if err != nil {
		app.notFound(w)
		return
	}

	forUser, err := app.user.GetFull(forUserID)
	page := forUser.Username + "'s comments"

	comments, err := app.comments.ForUser(user.ID, lastID, forUser.ID)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	var moreURL string
	if len(comments) == 30 {
		moreURL = strconv.Itoa(comments[len(comments)-1].ID)
	}

	app.render(w, r, "comments.page.tmpl", &templateData{
		Page:     page,
		Title:    strings.Title(page),
		Comments: comments,
		// StartRank: pageNum*30 + 1,
		MoreURL: moreURL,
	})
}

func (app *application) createPostForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		Page:  "submit",
		Title: "Submit",
		Form:  forms.New(nil),
	})
}

func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	user := app.authenticatedUser(r)

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title")
	form.MaxLength("title", 80)
	if form.Get("url") == "" {
		form.Required("text")
	} else {
		form.IsURL("url")
	}

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{
			Page:  "submit",
			Title: "Submit",
			Form:  form,
		})
		return
	}

	now := time.Now().UTC()
	postID, err := app.posts.Insert(mysql.CreatePostParams{
		Title:   form.Get("title"),
		Url:     form.Get("url"),
		Content: form.Get("text"),
		UserID:  user.ID,
		Created: now,
	})
	if err != nil {
		switch err {

		// Add upvote instead of creating new post.
		case models.ErrDuplicatePost:
			// TODO: potential race condition if it is possible to delete posts.
			// But undecided if that should be allowed or not.
			postID, err = app.posts.GetIDFromURL(form.Get("url"))
			if err != nil {
				app.serverError(w, err)
				return
			}
			_, err = app.posts.Upvote(postID, user.ID, now)
			if err != nil && err != models.ErrDuplicateUpvote {
				app.serverError(w, err)
				return
			}

		default:
			app.serverError(w, err)
			return
		}
	}
	// Add upvote for newly created post by user.
	_, err = app.posts.Upvote(postID, user.ID, now)
	if err != nil && err != models.ErrDuplicateUpvote {
		app.serverError(w, err)
		return
	}

	if form.Get("url") != "" && form.Get("text") != "" {
		// Add text as comment if post contained both URL and text.
		commentID, err := app.comments.Insert(mysql.CreateCommentParams{
			Content:  form.Get("text"),
			ParentID: 0,
			UserID:   user.ID,
			PostID:   postID,
			Created:  now,
		})
		if err != nil {
			app.serverError(w, err)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/items/%d#%d", postID, commentID), http.StatusSeeOther)
	}

	http.Redirect(w, r, fmt.Sprintf("/items/%d", postID), http.StatusSeeOther)
}

func (app *application) createReplyForm(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	postID, err := strconv.Atoi(params.ByName("postid"))
	if err != nil || postID < 1 {
		app.notFound(w)
		return
	}

	post, err := app.posts.GetByID(user.ID, postID)
	if err != nil || post.Deleted == 1 {
		app.notFound(w)
		return
	}

	parentID, err := strconv.Atoi(params.ByName("commentid"))
	if err != nil || parentID < 1 {
		app.notFound(w)
		return
	}

	parent, err := app.comments.GetByID(user.ID, parentID)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	replies, err := app.comments.ForComment(user.ID, postID, parentID)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	f := forms.New(make(url.Values))
	// Populate the comment-form with text if saved in the session (due to
	// redirect from login)
	f.Add("text", app.session.PopString(r, "text"))

	app.render(w, r, "reply.page.tmpl", &templateData{
		Page:     "reply",
		Title:    "Add Comment",
		Comment:  parent,
		Comments: replies,
		Form:     f,
	})
}

func (app *application) createReply(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	postID, err := strconv.Atoi(params.ByName("postid"))
	if err != nil || postID < 1 {
		app.notFound(w)
		return
	}

	post, err := app.posts.GetByID(user.ID, postID)
	if err != nil || post.Deleted == 1 {
		app.notFound(w)
		return
	}

	// Set the redirect URI to the post-site.
	redirect := fmt.Sprintf("/items/%d", postID)

	// The commentID from the URI denotes the parent ID for the comment reply.
	parentID, err := strconv.Atoi(params.ByName("commentid"))
	if err != nil || parentID < 0 {
		app.notFound(w)
		return
	}

	parent := &models.Comment{}
	if parentID != 0 {
		// If the parentID for the comment is set, the redirect URI is for
		// the comment reply site instead.
		redirect = fmt.Sprintf("/items/%d/comments/%d", postID, parentID)

		parent, err = app.comments.GetByID(user.ID, parentID)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("text")
	app.session.PopString(r, "text")

	if !form.Valid() {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}

	if parent.Deleted == 1 {
		app.session.Put(r, "flash", "You cannot reply to a deleted comment!")
		app.session.Put(r, "text", form.Get("text"))
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}

	id, err := app.comments.Insert(mysql.CreateCommentParams{
		Content:  form.Get("text"),
		ParentID: parentID,
		UserID:   user.ID,
		PostID:   postID,
		Created:  time.Now(),
	})
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s#%d", redirect, id), http.StatusSeeOther)
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Title: "Signup",
		Form:  forms.New(nil),
	})
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("username", "email", "password", "password2")
	form.MatchesPattern("username", forms.UsernameRX, "Username not allowed. Usernames can contain letters (a-z), numbers (0-9), and periods (.).")
	for _, v := range forms.RestrictedUsernames {
		if strings.Contains(strings.ToLower(form.Get("username")), v) {
			form.Errors.Add("username", "Username not allowed.")
			break
		}
	}
	form.MinLength("username", 2)
	form.MaxLength("username", 20)
	form.MatchesPattern("email", forms.EmailRX, "email is invalid")
	// Make sure we ban any @DOMAIN addresses.
	banned := strings.Split(form.Get("email"), "@")
	if strings.Contains(strings.ToLower(banned[len(banned)-1]), app.config.Server.Domain) { // Domain-level
		form.Errors.Add("email", "This email address is banned.")
	}
	form.MinLength("password", 8)
	if form.Valid() && (form.Get("password") != form.Get("password2")) {
		form.Errors.Add("password2", "Those passwords didn't match. Try again.")
	}

	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{
			Title: "Signup",
			Form:  form,
		})
		return
	}

	id, err := app.user.Insert(mysql.CreateUserParams{
		Username: form.Get("username"),
		Email:    form.Get("email"),
		Password: form.Get("password"),
	})
	if err != nil {
		switch err {

		case models.ErrDuplicateEmail:
			form.Errors.Add("email", "Email address already in use")
			app.render(w, r, "signup.page.tmpl", &templateData{
				Title: "Signup",
				Form:  form,
			})
			return

		case models.ErrDuplicateUsername:
			form.Errors.Add("username", "Username already in use")
			app.render(w, r, "signup.page.tmpl", &templateData{
				Title: "Signup",
				Form:  form,
			})
			return

		default:
			app.serverError(w, err)
			return

		}
	}

	app.session.Put(r, "flash", "Your signup was successful!")
	app.session.Put(r, "authenticatedUserID", id)

	http.Redirect(w, r, "/"+app.session.PopString(r, "goto"), http.StatusSeeOther)
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	form := forms.New(url.Values{})
	form.Add("goto", r.URL.Query().Get("goto"))

	app.render(w, r, "login.page.tmpl", &templateData{
		Page:  "login",
		Title: "Login",
		Form:  form,
	})
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)

	id, err := app.user.Authenticate(form.Get("username"), form.Get("password"))
	if err == models.ErrInvalidCredentials {
		app.session.Put(r, "flash", "Email or Password is incorrect")
		app.render(w, r, "login.page.tmpl", &templateData{
			Page:       "login",
			Title:      "Login",
			Form:       form,
			CurrentURL: form.Get("goto"),
			// CurrentURL: app.session.GetString(r, "goto"),
		})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "authenticatedUserID", id)

	http.Redirect(w, r, "/"+form.Get("goto"), http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "authenticatedUserID")

	app.session.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", 303)
}

func (app *application) upvotePost(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	postID, err := strconv.Atoi(params.ByName("postid"))
	if err != nil || postID < 1 {
		app.notFound(w)
		return
	}

	_, err = app.posts.Upvote(postID, user.ID, time.Now())
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/"+app.session.PopString(r, "goto"), http.StatusSeeOther)
}

func (app *application) upvoteComment(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	user := app.authenticatedUser(r)

	commentID, err := strconv.Atoi(params.ByName("commentid"))
	if err != nil || commentID < 1 {
		app.notFound(w)
		return
	}

	_, err = app.comments.Upvote(commentID, user.ID, time.Now())
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/"+app.session.PopString(r, "goto"), http.StatusSeeOther)
}

func (app *application) showProfile(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(contextParams).(httprouter.Params)
	id, err := app.user.GetIDByUsername(params.ByName("username"))
	if err != nil {
		app.notFound(w)
		return
	}

	u, err := app.user.GetFull(id)
	if err != nil {
		app.notFound(w)
		return
	}

	app.render(w, r, "profile.page.tmpl", &templateData{
		Title:       strings.Title(u.Username) + "'s Profile",
		ProfileUser: u,
	})
}

func (app *application) editProfileForm(w http.ResponseWriter, r *http.Request) {
	// params := r.Context().Value(contextParam).(httprouter.Params)
	// u, err := app.user.GetFromName(params.ByName("username"))
	// if err != nil {
	// 	app.notFound(w)
	// 	return
	// }

	// userID := app.session.GetInt(r, "authenticatedUserID")
	// if u.ID != userID {
	// 	app.notFound(w)
	// 	return
	// }

	app.render(w, r, "edit_profile.page.tmpl", &templateData{
		Title: "Edit Profile",
		Form:  forms.New(nil),
	})
}

// TODO: Changin email doesn't work
func (app *application) editProfile(w http.ResponseWriter, r *http.Request) {
	// params := r.Context().Value(contextParam).(httprouter.Params)
	user := app.authenticatedUser(r)

	// u, err := app.user.GetFromName(params.ByName("username"))
	// if err != nil {
	// 	app.notFound(w)
	// 	return
	// }

	// userID := app.session.GetInt(r, "authenticatedUserID")
	// if u.ID != userID {
	// 	app.notFound(w)
	// 	return
	// }

	// err = r.ParseForm()
	// if err != nil {
	// 	app.clientError(w, http.StatusBadRequest)
	// 	return
	// }

	form := forms.New(r.PostForm)
	form.Required("email")
	form.MatchesPattern("email", forms.EmailRX, "email is invalid.")
	form.MaxLength("aboutme", 455)

	if !form.Valid() {
		app.render(w, r, "edit_profile.page.tmpl", &templateData{
			Title: "Edit Profile",
			Form:  form,
		})
		return
	}

	err := app.user.Update(mysql.UpdateUserParams{
		ID:    user.ID,
		About: form.Get("aboutme"),
	})
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Profile updated!")

	http.Redirect(w, r, "/"+app.session.PopString(r, "goto"), http.StatusSeeOther)
}

func (app *application) guidelines(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "guidelines.page.tmpl", &templateData{Title: "Guidelines"})
}

func (app *application) acknowledgements(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "acknowledgements.page.tmpl", &templateData{Title: "Acknowledgements"})
}
