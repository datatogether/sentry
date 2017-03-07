package main

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// middleware handles request logging
func middleware(handler httprouter.Handle) httprouter.Handle {
	// no-auth middware func
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// poor man's logging:
		fmt.Println(r.Method, r.URL.Path, time.Now())

		// TODO - Strict Transport config?
		// if cfg.TLS {
		// 	// If TLS is enabled, set 1 week strict TLS, 1 week for now to prevent catastrophic mess-ups
		// 	w.Header().Add("Strict-Transport-Security", "max-age=604800")
		// }
		handler(w, r, p)
	}
}

// authMiddleware adds http basic auth if configured
func authMiddleware(handler httprouter.Handle) httprouter.Handle {
	// return auth middleware if configuration settings are present
	if cfg.HttpAuthUsername != "" && cfg.HttpAuthPassword != "" {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			// poor man's logging:
			fmt.Println(r.Method, r.URL.Path, time.Now())

			user, pass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(cfg.HttpAuthUsername)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(cfg.HttpAuthPassword)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Please enter your username and password for this site"`)
				w.WriteHeader(http.StatusUnauthorized)
				renderTemplate(w, "accessDenied.html")
				return
			}

			// TODO - Strict Transport config?
			// if cfg.TLS {
			// 	// If TLS is enabled, set 1 week strict TLS, 1 week for now to prevent catastrophic mess-ups
			// 	w.Header().Add("Strict-Transport-Security", "max-age=604800")
			// }
			handler(w, r, p)
		}
	}

	// no-auth middware func
	return middleware(handler)
}
