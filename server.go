package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

var (
	// cfg is the global configuration for the server. It's read in at startup from
	// the config.json file and enviornment variables, see config.go for more info.
	cfg *config

	// When was the last alert sent out?
	// Use this value to avoid bombing alerts
	lastAlertSent *time.Time

	// log output
	logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	// application database connection
	appDB *sql.DB
)

func main() {
	var err error
	cfg, err = initConfig(os.Getenv("GOLANG_ENV"))
	if err != nil {
		// panic if the server is missing a vital configuration detail
		panic(fmt.Errorf("server configuration error: %s", err.Error()))
	}

	connectToAppDb()

	if cfg.Crawl {
		// what a wonderful phrase :)
		go startCrawling()
	}

	// initialize a router to handle requests
	r := httprouter.New()

	// home handler, wrapped in middlware func
	r.GET("/", middleware(HandleDomains))

	// Seed a url to the crawler
	r.POST("/seed", middleware(SeedUrlHandler))

	// List domains
	r.GET("/domains", middleware(ListDomainsHandler))
	// Add a crawling domain
	r.POST("/domains", middleware(AddDomainHandler))

	r.GET("/urls", middleware(UrlsViewHandler))
	r.GET("/url", middleware(UrlMetadataHandler))
	r.POST("/url/meta", middleware(UrlSetMetadataHandler))

	r.POST("/context", middleware(SaveUrlContextHandler))
	r.DELETE("/context", middleware(DeleteUrlContextHandler))

	r.GET("/mem", middleware(MemStatsHandler))
	r.GET("/que", middleware(EnquedHandler))
	r.POST("/que", middleware(EnqueUrlHandler))
	r.POST("/shutdown", middleware(ShutdownHandler))

	r.POST("/archive", middleware(ArchiveUrlHandler))

	// serve static content from public directories
	r.ServeFiles("/css/*filepath", http.Dir("public/css"))
	r.ServeFiles("/js/*filepath", http.Dir("public/js"))

	// print notable config settings
	printConfigInfo()

	// fire it up!
	fmt.Println("starting server on port", cfg.Port)
	// non-tls configured servers end here
	if !cfg.TLS {
		logger.Fatal(StartHttpServer(cfg.Port, r))
	}
	// start server wrapped in a log.Fatal b/c http.ListenAndServe will not
	// return unless there's an error
	logger.Fatal(StartHttpsServer(cfg.Port, r))
}
