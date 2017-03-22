package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func StartServer(c *config, s *http.Server) error {
	s.Addr = fmt.Sprintf(fmt.Sprintf(":%s", c.Port))

	if !c.TLS {
		return s.ListenAndServe()
	}

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

	s.TLSConfig = &tls.Config{
		GetCertificate: certManager.GetCertificate,
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
	}

	return s.ListenAndServeTLS(cert, key)
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
