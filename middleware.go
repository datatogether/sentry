package main

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"
)

// middleware handles request logging
func middleware(handler http.HandlerFunc) http.HandlerFunc {
	// no-auth middware func
	return func(w http.ResponseWriter, r *http.Request) {
		// poor man's logging:
		fmt.Println(r.Method, r.URL.Path, time.Now())

		// If this server is operating behind a proxy, but we still want to force
		// users to use https, cfg.ProxyForceHttps == true will listen for the common
		// X-Forward-Proto & redirect to https
		if cfg.ProxyForceHttps {
			if r.Header.Get("X-Forwarded-Proto") == "http" {
				w.Header().Set("Connection", "close")
				url := "https://" + r.Host + r.URL.String()
				http.Redirect(w, r, url, http.StatusMovedPermanently)
				return
			}
		}

		// TODO - Strict Transport config?
		// if cfg.TLS {
		// 	// If TLS is enabled, set 1 week strict TLS, 1 week for now to prevent catastrophic mess-ups
		// 	w.Header().Add("Strict-Transport-Security", "max-age=604800")
		// }
		handler(w, r)
	}
}

// authMiddleware adds http basic auth if configured
func authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	// return auth middleware if configuration settings are present
	if cfg.HttpAuthUsername != "" && cfg.HttpAuthPassword != "" {
		return func(w http.ResponseWriter, r *http.Request) {
			// poor man's logging:
			fmt.Println(r.Method, r.URL.Path, time.Now())

			// If this server is operating behind a proxy, but we still want to force
			// users to use https, cfg.ProxyForceHttps == true will listen for the common
			// X-Forward-Proto & redirect to https
			if cfg.ProxyForceHttps {
				if r.Header.Get("X-Forwarded-Proto") == "http" {
					w.Header().Set("Connection", "close")
					url := "https://" + r.Host + r.URL.String()
					http.Redirect(w, r, url, http.StatusMovedPermanently)
					return
				}
			}

			user, pass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(cfg.HttpAuthUsername)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.HttpAuthPassword)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Please enter your username and password for this site"`)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("access denied \n"))
				return
			}

			// TODO - Strict Transport config?
			// if cfg.TLS {
			// 	// If TLS is enabled, set 1 week strict TLS, 1 week for now to prevent catastrophic mess-ups
			// 	w.Header().Add("Strict-Transport-Security", "max-age=604800")
			// }
			handler(w, r)
		}
	}

	// no-auth middware func
	return middleware(handler)
}
