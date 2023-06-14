package main

import (
	"database/sql"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/AnonymFromInternet/Purchases/internal/driver"
	"github.com/AnonymFromInternet/Purchases/internal/models"
	"github.com/alexedwards/scs/v2"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// For Frontend Binary

const version = "1.0.0"
const cssVersion = "1"

var SessionManager *scs.SessionManager

type config struct {
	port int
	env  string
	api  string

	db struct {
		dataSourceName string
	}

	stripe struct {
		secretKey string
		publicKey string
	}
	secretKeyForPasswordReset   string
	frontendURLForPasswordReset string
}

type application struct {
	config         config
	infoLog        *log.Logger
	errorLog       *log.Logger
	templateCache  map[string]*template.Template
	version        string
	DB             models.DBModel
	SessionManager *scs.SessionManager
}

func (application *application) serve() error {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", application.config.port),
		Handler:           application.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	application.infoLog.Println(fmt.Sprintf("Starting HTTP server in %s mode on port %d", application.config.env, application.config.port))

	return server.ListenAndServe()
}

func getConfig() config {
	var config config

	setAndParseFlags(&config)

	config.stripe.publicKey = os.Getenv("STRIPE_PUBLIC_KEY")
	config.stripe.secretKey = os.Getenv("STRIPE_SECRET_KEY")

	return config
}

func setAndParseFlags(config *config) {
	flag.IntVar(&config.port, "port", 4000, "Server port to listen on")
	flag.StringVar(&config.env, "env", "development", "Application environment{development|production}")
	flag.StringVar(&config.api, "api", "http://localhost:4001", "URL to api")
	flag.StringVar(&config.db.dataSourceName, "dsn", "host=localhost port=5432 dbname=postgres user=postgres password=", "DSN")
	flag.StringVar(&config.secretKeyForPasswordReset, "secretkey", "reset", "smtp password")
	flag.StringVar(&config.frontendURLForPasswordReset, "frontendURL", "http://localhost:4000", "frontend URL")

	flag.Parse()
}

func main() {
	gob.Register(models.TransactionData{})

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	config := getConfig()
	dbConn, err := driver.OpenDB(config.db.dataSourceName)
	if err != nil {
		errorLog.Fatal("cannot connect to the database: ", err)
	}

	defer func(dbConn *sql.DB) {
		err := dbConn.Close()
		if err != nil {
			errorLog.Fatal(err)
		}
	}(dbConn)

	SessionManager = scs.New()
	SessionManager.Lifetime = 24 * time.Hour

	templateCache := make(map[string]*template.Template)

	application := &application{
		config:         config,
		infoLog:        infoLog,
		errorLog:       errorLog,
		templateCache:  templateCache,
		version:        version,
		DB:             models.DBModel{DB: dbConn},
		SessionManager: SessionManager,
	}

	err = application.serve()
	if err != nil {
		application.errorLog.Println("cannot serve the app", err)
		log.Fatal(err)
	}
}
