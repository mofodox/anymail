package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"

	"github.com/mofodox/anymail/internal/database"
	"github.com/mofodox/anymail/internal/smtp"
	"github.com/mofodox/anymail/internal/version"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)

	err := run(logger)
	if err != nil {
		logger.Fatalf("%s\n%s", err, debug.Stack())
	}
}

type config struct {
	baseURL  string
	httpPort int
	db       struct {
		dsn         string
		automigrate bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		from     string
	}
	version bool
}

type application struct {
	config config
	db     *database.DB
	logger *log.Logger
	mailer *smtp.Mailer
	wg     sync.WaitGroup
}

func run(logger *log.Logger) error {
	var cfg config

	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:4444", "base URL for the application")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "user:pass@localhost:5432/db", "postgreSQL DSN")
	flag.BoolVar(&cfg.db.automigrate, "db-automigrate", true, "run migrations on startup")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "example.smtp.host", "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "smtp port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "example_username", "smtp username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "pa55word", "smtp password")
	flag.StringVar(&cfg.smtp.from, "smtp-from", "Example Name <no-reply@example.org>", "smtp sender")
	flag.BoolVar(&cfg.version, "version", false, "display version and exit")

	flag.Parse()

	if cfg.version {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.db.dsn, cfg.db.automigrate)
	if err != nil {
		return err
	}
	defer db.Close()

	mailer := smtp.NewMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.from)

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
		mailer: mailer,
	}

	return app.serveHTTP()
}
