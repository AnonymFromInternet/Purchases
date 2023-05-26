package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// For Frontend Binary

const version = "1.0.0"
const cssVersion = "1"

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
}

type application struct {
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	version       string
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

	flag.Parse()
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	templateCache := make(map[string]*template.Template)

	application := &application{
		config:        getConfig(),
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: templateCache,
		version:       version,
	}

	err := application.serve()
	if err != nil {
		application.errorLog.Println("[Frontend]:[main]:[main] - err := application.serve()", err)
		log.Fatal(err)
	}
}
