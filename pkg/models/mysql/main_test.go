package mysql_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/bjarnemagnussen/go-hackernews/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	mysql_migrate "github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/ory/dockertest/v3"
)

type dockerDBConn struct {
	Conn *sql.DB
}

var (
	// DockerDBConn holds the connection to our DB in the container we spin up for testing.
	DockerDBConn *dockerDBConn
	userModel    *mysql.UserModel
	postModel    *mysql.PostModel
	commentModel *mysql.CommentModel
)

func TestMain(m *testing.M) {
	pool, resource := initDB()
	resource.Expire(180)

	var err error
	userModel = &mysql.UserModel{DockerDBConn.Conn}
	postModel = &mysql.PostModel{DockerDBConn.Conn}
	commentModel, err = mysql.NewCommentModel(DockerDBConn.Conn)
	if err != nil {
		log.Fatalf("Could not create the comment model: %s", err)
	}

	code := m.Run()
	closeDB(pool, resource)
	os.Exit(code)
}

func initDB() (*dockertest.Pool, *dockertest.Resource) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	pool.MaxWait = time.Minute * 2
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mysql", "8", []string{"MYSQL_ROOT_PASSWORD=secret", "MYSQL_DATABASE=dbname"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	DockerDBConn = &dockerDBConn{}
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		DockerDBConn.Conn, err = sql.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/dbname?multiStatements=true&parseTime=true", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return DockerDBConn.Conn.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	DockerDBConn.initMigrations()

	return pool, resource

}

func closeDB(pool *dockertest.Pool, resource *dockertest.Resource) {
	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (db dockerDBConn) initMigrations() {
	driver, err := mysql_migrate.WithInstance(db.Conn, &mysql_migrate.Config{})
	if err != nil {
		log.Fatal(err)
	}

	migrate, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"dbname", driver)
	if err != nil {
		log.Fatal(err)
	}

	err = migrate.Up()
	if err != nil {
		log.Fatal(err)
	}
}
