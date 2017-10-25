package main

import (
	"database/sql"
	"fmt"
	"github.com/datatogether/core"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	_ "github.com/lib/pq"
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
	appDB = &sql.DB{}
	// hoist default store
	store = sql_datastore.DefaultStore
)

func init() {
	log.Out = os.Stdout
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}
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

	sqlutil.ConnectToDb("postgres", cfg.PostgresDbUrl, appDB)
	sql_datastore.SetDB(appDB)
	sql_datastore.Register(
		&core.Url{},
		&core.Link{},
	)

	// create any tables if they don't exist
	sc, err := sqlutil.LoadSchemaCommands(packagePath("sql/schema.sql"))
	if err != nil {
		log.Infof("error loading schema file: %s", err)
	} else {
		created, err := sc.Create(appDB, "primers", "sources", "urls", "links", "metadata", "snapshots", "collections")
		if err != nil {
			log.Infof("error creating missing tables: %s", err)
		} else if len(created) > 0 {
			log.Info("created tables:", created)
		}
	}

	// always crawl seeds
	go startCrawlingSeeds()

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

	// start server wrapped in a log.Fatal. http.ListenAndServe will not
	// return unless a fatal error occurs
	log.Fatal(StartServer(cfg, s))
}

// NewServerRoutes returns a Muxer that has all API routes.
// This makes for easy testing using httptest, see server_test.go
func NewServerRoutes() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/.well-known/acme-challenge/", CertbotHandler)
	m.Handle("/", middleware(HealthCheckHandler))
	m.Handle("/healthcheck", middleware(HealthCheckHandler))

	// Seed a url to the crawler
	// r.POST("/seed", middleware(SeedUrlHandler))

	// List domains
	// m.Handle("/primers", middleware(ListPrimersHandler))
	// Add a crawling domain
	// r.POST("/primers", middleware(AddPrimerHandler))

	m.Handle("/urls", middleware(UrlsHandler))
	// m.Handle("/url", middleware(UrlHandler))
	m.Handle("/sources", middleware(CrawlingSourcesHandler))
	m.Handle("/mem", middleware(MemStatsHandler))
	m.Handle("/que", middleware(QueHandler))
	m.Handle("/shutdown", middleware(ShutdownHandler))

	return m
}
