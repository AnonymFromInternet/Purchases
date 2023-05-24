package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// For Backend Binary

const version = "1.0.0"

type config struct {
	port int
	env  string

	db struct {
		dataSourceName string
	}

	stripe struct {
		secretKey string
		publicKey string
	}
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
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

	application.infoLog.Println(fmt.Sprintf("Starting Backend server in %s mode on port %d", application.config.env, application.config.port))

	return server.ListenAndServe()
}

func setConfig(config *config) {
	flag.IntVar(&config.port, "port", 4001, "Server port to listen on")
	flag.StringVar(&config.env, "env", "development", "Application environment{development|production|maintenance}")

	flag.Parse()

	config.stripe.publicKey = os.Getenv("STRIPE_PUBLIC_KEY")
	config.stripe.secretKey = os.Getenv("STRIPE_SECRET_KEY")
}

func main() {
	var config config
	setConfig(&config)

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	application := &application{
		config:   config,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
	}

	err := application.serve()
	if err != nil {
		application.errorLog.Println("[Backend]:[main]:[main] - err := application.serve()", err)
		log.Fatal(err)
	}

}
