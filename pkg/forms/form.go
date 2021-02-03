package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// EmailRX follows the W3C recommendation:
// https://www.w3.org/TR/2016/REC-html51-20161101/sec-forms.html#email-state-typeemail
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// UsernameRX allows only letters and numbers and dots (which are ignored):
var UsernameRX = regexp.MustCompile("^(([a-z]|[A-Z]|[0-9])+(\\.?))+$")

// RestrictedUsernames contains reserved usernames (pre-defined list of special banned and reserved keywords in names,
// such as "root", "www", "admin")
var RestrictedUsernames = []string{
	"about", "access", "account", "accounts", "add", "address", "adm", "admin", "administration", "adult",
	"advertising", "affiliate", "affiliates", "ajax", "analytics", "android", "anon", "anonymous", "api", "app", "apps",
	"archive", "atom", "auth", "authentication", "avatar", "backup", "banner", "banners", "bin", "billing", "blog",
	"blogs", "board", "bot", "bots", "business", "chat", "cache", "cadastro", "calendar", "campaign", "careers", "cgi",
	"client", "cliente", "code", "comercial", "compare", "config", "connect", "contact", "contest", "create", "code",
	"compras", "css", "dashboard", "data", "db", "design", "delete", "demo", "design", "designer", "dev", "devel",
	"dir", "directory", "doc", "docs", "domain", "download", "downloads", "edit", "editor", "email", "ecommerce",
	"forum", "forums", "faq", "favorite", "feed", "feedback", "flog", "follow", "file", "files", "free", "ftp",
	"gadget", "gadgets", "games", "guest", "group", "groups", "help", "home", "homepage", "host", "hosting", "hostname",
	"html", "http", "httpd", "https", "hpg", "info", "information", "image", "img", "images", "imap", "index", "invite",
	"intranet", "indice", "ipad", "iphone", "irc", "java", "javascript", "job", "jobs", "js", "knowledgebase", "log",
	"login", "logs", "logout", "list", "lists", "mail", "mail1", "mail2", "mail3", "mail4", "mail5", "mailer",
	"mailing", "mx", "manager", "marketing", "master", "me", "media", "message", "microblog", "microblogs", "mine",
	"mp3", "msg", "msn", "mysql", "messenger", "mob", "mobile", "movie", "movies", "music", "musicas", "my", "name",
	"named", "net", "network", "new", "news", "newsletter", "nick", "nickname", "notes", "noticias", "ns", "ns1", "ns2",
	"ns3", "ns4", "old", "online", "operator", "order", "orders", "page", "pager", "pages", "panel", "password", "perl",
	"pic", "pics", "photo", "photos", "photoalbum", "php", "plugin", "plugins", "pop", "pop3", "post", "postmaster",
	"postfix", "posts", "profile", "project", "projects", "promo", "pub", "public", "python", "random", "register",
	"registration", "root", "ruby", "rss", "sale", "sales", "sample", "samples", "script", "scripts", "secure", "send",
	"service", "shop", "sql", "signup", "signin", "search", "security", "settings", "setting", "setup", "site", "sites",
	"sitemap", "smtp", "soporte", "ssh", "stage", "staging", "start", "subscribe", "subdomain", "suporte", "support",
	"stat", "static", "stats", "status", "store", "stores", "system", "tablet", "tablets", "tech", "telnet", "test",
	"test1", "test2", "test3", "teste", "tests", "theme", "themes", "tmp", "todo", "task", "tasks", "tools", "tv",
	"talk", "update", "upload", "url", "user", "username", "usuario", "usage", "vendas", "video", "videos", "visitor",
	"win", "ww", "www", "www1", "www2", "www3", "www4", "www5", "www6", "www7", "wwww", "wws", "wwws", "web",
	"webmail", "website", "websites", "webmaster", "workshop", "xxx", "xpg", "you", "yourname", "yourusername",
	"yoursite", "yourdomain", "whoishiring",

	"anal", "anus", "arse", "ass", "ballsack", "balls", "bastard", "bitch", "biatch", "bloody", "blowjob", "bollock",
	"bollok", "bonerboob", "bugger", "bum", "butt", "buttplug", "clitoris", "cock", "coon", "crap", "cunt", "damn",
	"dick", "dildo", "dyke", "fag", "feckfellate", "fellatio", "felching", "fuck", "fudgepacker", "fudge", "packer",
	"flange", "Goddamn", "God", "damn", "hell", "homojerk", "jizz", "knobend", "knob", "end", "labia", "lmao", "lmfao",
	"muff", "nigger", "nigga", "omg", "penis", "piss", "poop", "prick", "pubepussy", "queer", "scrotum", "sex", "shit",
	"sh1t", "slut", "smegma", "spunk", "tit", "tosser", "turd", "twat", "vagina", "wank", "whore", "wtf",
}

// Create a custom Form struct, which anonymously embeds a url.Values object
// (to hold the form data) and an Errors field to hold any validation errors
// for the form data.
type Form struct {
	url.Values
	Errors errors
}

// Define a New function to initialize a custom Form struct. Notice that
// this takes the form data as the parameter?
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// Implement a MaxLength method to check that a specific field in the form
// contains a maximum number of characters. If the check fails then add the
// appropriate message to the form errors.
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d))
	}
}

func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d characters)", d))
	}
}

func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" {
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This field is invalid")
}

func (f *Form) IsURL(field string) {
	value := f.Get(field)
	u, err := url.Parse(value)
	if err != nil {
		f.Errors.Add(field, "Must be a valid URL")
	} else if u.Scheme == "" || u.Host == "" {
		f.Errors.Add(field, "Must be an absolute URL")
	} else if u.Scheme != "http" && u.Scheme != "https" {
		f.Errors.Add(field, "Must begin with http or https")
	}
}

func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp, errorStr string) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, errorStr)
	}
}

// Implement a Valid method which returns true if there are no errors.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
