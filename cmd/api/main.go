package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/mofodox/anymail/internal/data"
	"github.com/mofodox/anymail/internal/database"
	"github.com/mofodox/anymail/internal/smtp"
	"github.com/mofodox/anymail/internal/version"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)

	if err := godotenv.Load(); err != nil {
		logger.Fatal(err)
	} else {
		logger.Println("successfully loaded .env file")
	}

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
	models data.Models
	mailer *smtp.Mailer
	wg     sync.WaitGroup
}

func run(logger *log.Logger) error {
	var cfg config

	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT_MAILTRAP"))

	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:4444", "base URL for the application")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "postgreSQL DSN")
	flag.BoolVar(&cfg.db.automigrate, "db-automigrate", true, "run migrations on startup")
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST_MAILTRAP"), "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", smtpPort, "smtp port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME_MAILTRAP"), "smtp username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD_MAILTRAP"), "smtp password")
	flag.StringVar(&cfg.smtp.from, "smtp-from", "MyTengah <no-reply@mytengah.sg>", "smtp sender")
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
		models: data.NewModels(db),
		mailer: mailer,
	}

	return app.serveHTTP()
}
