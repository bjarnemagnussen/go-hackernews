package main

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
)

var mockPostHTML = `<span class="title"><a target="_blank" href="https://google.com/ref/some_link">Post Test Title</a></span>`

// TODO: Use regular expressions to ensure that the ordering of comments is correct.
var mockCommentParentHTML = `This is a test comment.`
var mockCommentChildHTML = `This is a test reply.`

func TestShowPost(t *testing.T) {

	// Create a new instance of our application struct which uses the mocked
	// dependencies.
	app := newTestApplication(t)

	// Establish a new test server for running end-to-end tests.
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Set up some table-driven tests to check the responses sent by our
	// application for different URLs.
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody [][]byte
	}{
		{"Valid ID", "/items/1", http.StatusOK, [][]byte{[]byte(mockPostHTML), []byte(mockCommentParentHTML), []byte(mockCommentChildHTML)}},
		{"Non-existent ID", "/items/2", http.StatusNotFound, nil},
		{"Negative ID", "/items/-1", http.StatusNotFound, nil},
		{"Decimal ID", "/items/1.23", http.StatusNotFound, nil},
		{"String ID", "/items/foo", http.StatusNotFound, nil},
		{"Empty ID", "/items/", http.StatusNotFound, nil},
		{"Trailing slash", "/items/1/", http.StatusMovedPermanently, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			for _, b := range tt.wantBody {
				if !bytes.Contains(body, b) {
					t.Errorf("want body to contain %q", b)
				}
			}
		})
	}
}

func TestShowNewestComments(t *testing.T) {
	// Create a new instance of our application struct which uses the mocked
	// dependencies.
	app := newTestApplication(t)

	// Establish a new test server for running end-to-end tests.
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Set up some table-driven tests to check the responses sent by our
	// application for different URLs.
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody [][]byte
	}{
		{"Endpoint", "/items/1/comments/1", http.StatusOK, [][]byte{[]byte(mockCommentParentHTML), []byte(mockCommentChildHTML)}},
		{"Endpoint", "/items/1/comments/2", http.StatusOK, [][]byte{[]byte(mockCommentChildHTML)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			for _, b := range tt.wantBody {
				if !bytes.Contains(body, b) {
					t.Errorf("want body to contain %q", b)
				}
			}
		})
	}
}

func TestSignupUser(t *testing.T) {
	app := newTestApplication(t)
	app.config.Server.Domain = "mypage.com"
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, body)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid submission", "Bob", "bob@example.com", "validPa$$word", csrfToken, http.StatusSeeOther, nil},
		{"Empty name", "", "bob@example.com", "validPa$$word", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Empty email", "Bob", "", "validPa$$word", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Empty password", "Bob", "bob@example.com", "", csrfToken, http.StatusOK, []byte("This field cannot be blank")},
		{"Invalid email (incomplete domain)", "Bob", "bob@example.", "validPa$$word", csrfToken, http.StatusOK, []byte("email is invalid")},
		{"Invalid email (missing @)", "Bob", "bobexample.com", "validPa$$word", csrfToken, http.StatusOK, []byte("email is invalid")},
		{"Invalid email (missing local part)", "Bob", "@example.com", "validPa$$word", csrfToken, http.StatusOK, []byte("email is invalid")},
		{"Short password", "Bob", "abc@example.com", "pa$word", csrfToken, http.StatusOK, []byte("This field is too short (minimum is 8 characters)")},
		{"Duplicate email", "Bob", "dupe@example.com", "validPa$$word", csrfToken, http.StatusOK, []byte("Email address already in use")},
		{"Invalid CSRF Token", "", "", "", "wrongToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("username", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("password2", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}
