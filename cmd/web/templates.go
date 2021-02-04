package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/forms"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type templateData struct {
	CSRFToken       string
	Page            string
	Title           string
	Flash           string
	Form            *forms.Form
	IsAuthenticated bool
	User            *models.User
	ProfileUser     *models.User
	Post            *models.Post
	Posts           []*models.Post
	Comment         *models.Comment
	Comments        []*models.Comment
	StartRank       int
	CurrentURL      string
	MoreURL         string

	SiteName   string
	SiteShort  string
	Generator  string
	ThemeColor string
	FooterText string
}

var functions = template.FuncMap{

	// add adds the two provided integers.
	"add": func(a, b int) int {
		return a + b
	},

	// mul multiplies the two provided integers.
	"mul": func(a, b int) int {
		return a * b
	},

	// min returns the minimum for two integers
	"min": func(a, b int) int {
		if a < b {
			return a
		}
		return b
	},

	// min returns the minimum for two integers
	"is": func(s, substr string) bool {
		if strings.Contains(s, substr) {
			return true
		}
		return false
	},

	"markDown": markDowner,

	// humanDate converts the provided time.Time to a human-readable string.
	"humanDate": humanDate,

	// moment takes two time.Time and returns a string denoting the largest
	// unit of time difference.
	"moment": moment,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts

	}

	return cache, nil
}

func moment(t1, t2 time.Time) string {
	// 	year, month, day, hour, min, sec := timeDiff(t1, t2)
	diff := t2.Sub(t1)
	if years := diff.Hours() / (24 * 365.25); int(years) > 1 {
		return fmt.Sprintf("%d years", int(years))
	} else if diff.Hours()/(24*365.25) >= 1 {
		return "1 year"
	} else if months := diff.Hours() / (24 * 30.5); int(months) > 1 {
		return fmt.Sprintf("%d months", int(months))
	} else if diff.Hours()/(24*30.5) >= 1 {
		return "1 month"
	} else if days := diff.Hours() / 24; int(days) > 1 {
		return fmt.Sprintf("%d days", int(days))
	} else if diff.Hours()/24 >= 1 {
		return "1 day"
	} else if int(diff.Hours()) > 1 {
		return fmt.Sprintf("%d hours", int(diff.Hours()))
	} else if diff.Hours() >= 1 {
		return "1 hour"
	} else if int(diff.Minutes()) > 1 {
		return fmt.Sprintf("%d mins", int(diff.Minutes()))
	} else if diff.Minutes() >= 1 {
		return "1 min"
	}
	return fmt.Sprintf("%d secs", int(diff.Seconds()))
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func markDowner(args ...interface{}) string {
	// unsafe := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)), blackfriday.WithExtensions(blackfriday.CommonExtensions))
	unsafe := blackfriday.MarkdownCommon([]byte(fmt.Sprintf("%s", args...)))

	// Preserve classes of fenced code blocks while using the bluemonday HTML sanitizer.
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	html := p.SanitizeBytes(unsafe)

	return string(html)
}
