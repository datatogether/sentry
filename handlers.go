package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/julienschmidt/httprouter"
)

// templates is a collection of views for rendering with the renderTemplate function
// see homeHandler for an example
var templates = template.Must(template.ParseFiles("views/index.html", "views/expired.html", "views/accessDenied.html", "views/notFound.html"))

func MemStatsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mu.Lock()
	w.Write(memStats(nil))
	mu.Unlock()
}

func EnquedHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mu.Lock()
	w.Write(enquedDomains())
	mu.Unlock()
}

func reqUrl(r *http.Request) (*url.URL, error) {
	return url.Parse(r.FormValue("url"))
}

func SeedUrlHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if queue != nil {
		// u, err := NormalizeURLString(r.FormValue("url"))
		u, err := reqUrl(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, fmt.Sprintf("'%s' is not a valid url", r.FormValue("url")))
			return
		}
		queue.SendStringGet(u.String())
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("added url: %s", u.String()))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf("'%s' is not a valid url", r.FormValue("url")))
	}
}

func UrlMetadataHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqUrl, err := reqUrl(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf("'%s' is not a valid url", r.FormValue("url")))
		return
	}

	u := &Url{Url: reqUrl}
	if err := u.Read(appDB); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf("read url '%s' err: %s", reqUrl.String(), err.Error()))
		return
	}

	data, err := json.Marshal(u.Metadata())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf("encode json error: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func UrlAddMetadataHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// u, err := reqUrl(r)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	io.WriteString(w, fmt.Sprintf("'%s' is not a valid url", r.FormValue("url")))
	// 	return
	// }
}

func ShutdownHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	stopCrawler <- true
	w.Write([]byte("shutting down"))
}

// HomeHandler renders the home page
func HomeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	renderTemplate(w, "index.html")
}

// renderTemplate renders a template with the values of cfg.TemplateData
func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl, cfg.TemplateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleDomains(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rows, err := appDB.Query(fmt.Sprintf("select %s from domains", domainCols()))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	domains := []*Domain{}
	for rows.Next() {
		d := &Domain{}
		if err := d.UnmarshalSQL(rows); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		domains = append(domains, d)
	}

	json.NewEncoder(w).Encode(domains)
}

// middleware handles request logging, expiry & authentication if set
func middleware(handler httprouter.Handle) httprouter.Handle {
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

			handler(w, r, p)
		}
	}

	// no-auth middware func
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// poor man's logging:
		fmt.Println(r.Method, r.URL.Path, time.Now())
		handler(w, r, p)
	}
}
