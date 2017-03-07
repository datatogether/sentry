package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/acme/autocert"
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

// StartHttpServer fires up an http server on port, with a given handler
func StartHttpServer(port string, handler http.Handler) error {
	return http.ListenAndServe(fmt.Sprintf(":%s", port), handler)
}

// Start an HTTPS server on port
// worthwhile reading from Filippo: https://blog.cloudflare.com/exposing-go-on-the-internet/
func StartHttpsServer(port string, handler http.Handler) error {
	certCache := "/tmp/certs"
	key, cert := "", ""

	// LetsEncrypt is good. Thanks LetsEncrypt.
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.UrlRoot),
		Cache:      autocert.DirCache(certCache),
	}

	// Attempt to boot a port 80 https redirect
	go func() { HttpsRedirect() }()

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			// Causes servers to use Go's default ciphersuite preferences,
			// which are tuned to avoid attacks. Does nothing on clients.
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		},
	}
	server.Handler = handler
	return server.ListenAndServeTLS(cert, key)
}

// Redirect HTTP to https if port 80 is open
func HttpsRedirect() {
	ln, err := net.Listen("tcp", ":80")
	if err != nil {
		return
	}

	logger.Println("TCP Port 80 is available, redirecting traffic to https")

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			url := "https://" + req.Host + req.URL.String()
			http.Redirect(w, req, url, http.StatusMovedPermanently)
		}),
	}
	logger.Fatal(srv.Serve(ln))
}
