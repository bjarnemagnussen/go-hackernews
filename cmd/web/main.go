package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models"
	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
	"github.com/golangcollege/sessions"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"

	_ "github.com/go-sql-driver/mysql"
)

type contextKey string

var contextParams = contextKey("params")
var contextAuthenticatedUser = contextKey("authenticatedUser")

type config struct {
	Server struct {
		Port   string `yaml:"port" envconfig:"SERVER_PORT"`
		Host   string `yaml:"host" envconfig:"SERVER_HOST"`
		Domain string `yaml:"domain" envconfig:"SERVER_DOMAIN"`
		Secret string `yaml:"secret_key" envconfig:"SERVER_SECRET"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn" envconfig:"DATABASE_DSN"`
	} `yaml:"database"`
	Template struct {
		SiteName   string `yaml:"site_name" envconfig:"TEMPLATE_SITENAME"`
		SiteShort  string `yaml:"site_short" envconfig:"TEMPLATE_SITESHORT"`
		Generator  string `yaml:"generator" envconfig:"TEMPLATE_GENERATOR"`
		ThemeColor string `yaml:"theme_color" envconfig:"TEMPLATE_THEMECOLOR"`
		FooterText string `yaml:"footer_text" envconfig:"TEMPLATE_FOOTERTEXT"`
	} `yaml:"template"`
}

type application struct {
	config        *config
	infoLog       *log.Logger
	errorLog      *log.Logger
	session       *sessions.Session
	templateCache map[string]*template.Template

	// user contains the DB model and is defined as an interface to allow
	// mocking the model for testing purposes.
	user interface {
		Insert(arg mysql.CreateUserParams) (int, error)
		Update(arg mysql.UpdateUserParams) error
		Authenticate(username, password string) (int, error)
		GetByID(id int) (*models.User, error)
		GetIDByUsername(name string) (int, error)
		GetFull(id int) (*models.User, error)
	}

	// posts contains the DB model and is defined as an interface to allow
	// mocking the model for testing purposes.
	posts interface {
		Insert(arg mysql.CreatePostParams) (int, error)
		GetByID(loggedInUserID, id int) (*models.Post, error)
		GetIDFromURL(s string) (int, error)
		GetFromDomain(loggedInUserID, lastID int, domain string) ([]*models.Post, error)
		GetForType(loggedInUserID, lastID int, popular bool, postType mysql.PostType) ([]*models.Post, error)
		Latest(loggedInUserID, lastPostID, filterByUserID int) ([]*models.Post, error)
		Popular(loggedInUserID, offset int) ([]*models.Post, error)
		Upvote(postID, userID int, created time.Time) (int, error)
	}

	// comments contains the DB model and is defined as an interface to allow
	// mocking the model for testing purposes.
	comments interface {
		Insert(arg mysql.CreateCommentParams) (int, error)
		GetByID(loggedInUserID, id int) (*models.Comment, error)
		ForPost(loggedInUserID, postID int) ([]*models.Comment, error)
		ForUser(loggedInUserID, lastCommentID, id int) ([]*models.Comment, error)
		ForComment(userID, postID, commentID int) ([]*models.Comment, error)
		Latest(userID, lastID int) ([]*models.Comment, error)
		Upvote(commentID, userID int, created time.Time) (int, error)
	}
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var cfg config
	err := readFile(&cfg)
	if err != nil {
		errorLog.Fatal(err)
	}
	err = readEnv(&cfg)
	if err != nil {
		errorLog.Fatal(err)
	}

	db, err := openDB(cfg.Database.DSN + "?parseTime=true")
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	commentModel, err := mysql.NewCommentModel(db)
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(cfg.Server.Secret))
	session.Lifetime = 12 * time.Hour
	session.SameSite = http.SameSiteStrictMode
	// session.Secure = true // TODO: For production with SSL

	app := &application{
		config:        &cfg,
		infoLog:       infoLog,
		errorLog:      errorLog,
		posts:         &mysql.PostModel{DB: db},
		comments:      commentModel,
		session:       session,
		templateCache: templateCache,
		user:          &mysql.UserModel{DB: db},
	}

	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s\n", cfg.Server.Host+":"+cfg.Server.Port)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func readFile(cfg *config) error {
	f, err := os.Open("config.yml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		panic(err)
	}

	return nil
}

func readEnv(cfg *config) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return err
	}

	return nil
}
