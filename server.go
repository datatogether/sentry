package main

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

var (
	// cfg is the global configuration for the server. It's read in at startup from
	// the config.json file and enviornment variables, see config.go for more info.
	cfg *config

	// When was the last alert sent out?
	// Use this value to avoid bombing alerts
	// TODO - this is from an half-baked, unfinished alerts feature idea
	lastAlertSent *time.Time

	// log output
	// logger = logger.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	log = logrus.New()

	// application database connection
	appDB *sql.DB
)

func init() {
	log.Out = os.Stdout
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.TextFormatter{}
}

// NewServerRoutes returns a Muxer that has all API routes.
// This makes for easy testing using httptest, see server_test.go
func NewServerRoutes() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/.well-known/acme-challenge/", CertbotHandler)
	m.Handle("/", middleware(HealthCheckHandler))

	// Seed a url to the crawler
	// r.POST("/seed", middleware(SeedUrlHandler))

	// List domains
	// m.Handle("/primers", middleware(ListPrimersHandler))
	// Add a crawling domain
	// r.POST("/primers", middleware(AddPrimerHandler))

	m.Handle("/urls", middleware(UrlsHandler))
	// m.Handle("/url", middleware(UrlHandler))
	m.Handle("/mem", middleware(MemStatsHandler))
	m.Handle("/que", middleware(EnquedHandler))
	// r.POST("/que", middleware(EnqueUrlHandler))
	m.Handle("/shutdown", middleware(ShutdownHandler))

	return m
}

func main() {
	var err error
	cfg, err = initConfig(os.Getenv("GOLANG_ENV"))
	if err != nil {
		// panic if the server is missing a vital configuration detail
		panic(fmt.Errorf("server configuration error: %s", err.Error()))
	}
	if cfg.Debug {
		log.Level = logrus.DebugLevel
	}

	connectToAppDb()

	if cfg.Crawl {
		// what a wonderful phrase :)
		go startCrawling()
	}

	// run cron every 5 hours for now
	go StartCron(time.Hour * 5)

	s := &http.Server{}
	// connect mux to server
	s.Handler = NewServerRoutes()

	// print notable config settings
	// printConfigInfo()

	// fire it up!
	fmt.Println("starting server on port", cfg.Port)

	// start server wrapped in a log.Fatal b/c http.ListenAndServe will not
	// return unless there's an error
	log.Fatal(StartServer(cfg, s))
}
