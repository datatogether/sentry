package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

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
