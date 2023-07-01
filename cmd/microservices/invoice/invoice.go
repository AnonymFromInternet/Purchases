package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int

	smtp struct {
		host     string
		port     int
		username string
		password string
	}

	frontendURL string
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
}

func main() {
	var config config
	flag.IntVar(&config.port, "port", 8000, "Microservice server port to listen on")

	flag.StringVar(&config.smtp.host, "smtphost", "sandbox.smtp.mailtrap.io", "smtp host")
	flag.IntVar(&config.smtp.port, "smtpport", 587, "smtp port")
	flag.StringVar(&config.smtp.username, "smtpusername", "2bd6b713077f69", "smtp username")
	flag.StringVar(&config.smtp.password, "smtppassword", "9770e176b68323", "smtp password")
	flag.StringVar(&config.frontendURL, "frontendURL", "http://localhost:4000", "frontend URL")

	flag.Parse()

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
		application.errorLog.Println("[Invoice microservice]:[main]:[main] - err := application.serve()", err)
		log.Fatal(err)
	}
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

	application.infoLog.Println(fmt.Sprintf("Starting microservice on port %d", application.config.port))

	return server.ListenAndServe()
}
