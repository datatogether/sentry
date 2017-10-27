package main

import (
	"encoding/json"
	"fmt"
	"github.com/datatogether/core"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func reqParamInt(key string, r *http.Request) (int, error) {
	i, err := strconv.ParseInt(r.FormValue(key), 10, 0)
	return int(i), err
}

func reqParamBool(key string, r *http.Request) (bool, error) {
	return strconv.ParseBool(r.FormValue(key))
}

// HealthCheckHandler is a basic "hey I'm fine" for load balancers & co
// TODO - add Database connection & proper configuration checks here for more accurate
// health reporting
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{ "status" : 200 }`))
}

// CertbotHandler pipes the certbot response for manual certificate generation
func CertbotHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, cfg.CertbotResponse)
}

func MemStatsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	w.Write(memStats(nil))
	mu.Unlock()
}

func QueHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		mu.Lock()
		w.Write(enquedUrls())
		mu.Unlock()
	case "POST":
		SeedUrlHandler(w, r)
	default:
		NotFoundHandler(w, r)
	}
}

func reqUrl(r *http.Request) (*url.URL, error) {
	return url.Parse(r.FormValue("url"))
}

func SeedUrlHandler(w http.ResponseWriter, r *http.Request) {
	if seedQueue != nil {
		// u, err := NormalizeURLString(r.FormValue("url"))
		parsedUrl, err := reqUrl(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, fmt.Sprintf("'%s' is not a valid url", r.FormValue("url")))
			return
		}

		if err := core.ValidArchivingUrl(appDB, parsedUrl.String()); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		_, err = appDB.Exec("insert into core_requests (created,url,user_id) values ($1, $2, $3)", time.Now().Round(time.Second).In(time.UTC), parsedUrl.String(), "")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, fmt.Sprintf("save url error: %s", err.Error()))
			return
		}

		u := &core.Url{Url: parsedUrl.String()}
		if err := u.Read(store); err != nil {
			if err == core.ErrNotFound {
				if err := u.Save(store); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, fmt.Sprintf("save url error: %s", err.Error()))
					return
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, fmt.Sprintf("save url error: %s", err.Error()))
				return
			}
		}

		mu.Lock()
		defer mu.Unlock()
		enqued[parsedUrl.String()] = "GET"

		seedQueue.SendStringGet(parsedUrl.String())
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("added url: %s", parsedUrl.String()))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf("'%s' is not a valid url", r.FormValue("url")))
	}
}

// TODO - fix
// AddPrimerHandler adds a primer for crawling.
// func AddPrimerHandler(w http.ResponseWriter, r *http.Request) {
// 	parsed, err := url.Parse(r.FormValue("url"))
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("parse url '%s' error: %s", r.FormValue("url"), err.Error()))
// 		return
// 	}

// 	d := &Primer{
// 		Host:  parsed.Host,
// 		Crawl: true,
// 	}

// 	if err := d.Insert(appDB); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, err.Error())
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// }

// func UrlMetadataHandler(w http.ResponseWriter, r *http.Request) {
// 	reqUrl, err := reqUrl(r)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("'%s' is not a valid url", r.FormValue("url")))
// 		return
// 	}

// 	u := &core.Url{Url: reqUrl.String()}
// 	if err := u.Read(appDB); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("read url '%s' err: %s", reqUrl.String(), err.Error()))
// 		return
// 	}

// 	meta, err := u.Metadata(appDB)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		io.WriteString(w, fmt.Sprintf("read url '%s' err: %s", reqUrl.String(), err.Error()))
// 		return
// 	}

// 	data, err := json.MarshalIndent(meta, "", "  ")
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		io.WriteString(w, fmt.Sprintf("encode json error: %s", err.Error()))
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Header().Add("Content-Type", "application/json")
// 	w.Write(data)
// }

// func SaveUrlContextHandler(w http.ResponseWriter, r *http.Request) {
// 	uc := &UrlContext{}
// 	if err := json.NewDecoder(r.Body).Decode(uc); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("json formatting error: %s", err.Error()))
// 		return
// 	}
// 	r.Body.Close()

// 	if err := uc.Save(appDB); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("error saving context: %s", err.Error()))
// 		return
// 	}

// 	w.WriteHeader(200)
// 	w.Header().Add("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(uc); err != nil {
// 		log.Debug(err.Error())
// 	}
// }

// func DeleteUrlContextHandler(w http.ResponseWriter, r *http.Request) {
// 	uc := &UrlContext{}
// 	if err := json.NewDecoder(r.Body).Decode(uc); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("json formatting error: %s", err.Error()))
// 		return
// 	}
// 	r.Body.Close()

// 	if err := uc.Delete(appDB); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("error saving context: %s", err.Error()))
// 		return
// 	}

// 	w.WriteHeader(200)
// 	io.WriteString(w, "url deleted")
// }

// func UrlSetMetadataHandler(w http.ResponseWriter, r *http.Request) {
// 	reqUrl, err := reqUrl(r)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("'%s' is not a valid url", r.FormValue("url")))
// 		return
// 	}

// 	u := &Url{Url: reqUrl.String()}
// 	if err := u.Read(appDB); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("read url '%s' err: %s", reqUrl.String(), err.Error()))
// 		return
// 	}

// 	defer r.Body.Close()
// 	meta := []interface{}{}
// 	if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		io.WriteString(w, fmt.Sprintf("json parse err: %s", err.Error()))
// 		return
// 	}
// 	u.Meta = meta

// 	if err := u.Update(appDB); err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		io.WriteString(w, fmt.Sprintf("save url error: %s", err.Error()))
// 		return
// 	}

// 	m, err := u.Metadata(appDB)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		io.WriteString(w, fmt.Sprintf("url metadata error: %s", err.Error()))
// 		return
// 	}
// 	data, err := json.Marshal(m)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		io.WriteString(w, fmt.Sprintf("encode json error: %s", err.Error()))
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Header().Add("Content-Type", "application/json")
// 	w.Write(data)
// }

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found\n"))
}

func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		stopCrawler <- true
		w.Write([]byte("shutting down\n"))
	default:
		NotFoundHandler(w, r)
	}
}

// HomeHandler renders the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hi there!")
}

func UrlsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// if we have a "url" param, read that single url
		url := r.FormValue("url")
		if url != "" {
			u := &core.Url{Url: url}
			if err := u.Read(store); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Debug(err.Error())
				return
			}

			data, err := json.MarshalIndent(u, "", "  ")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Debug(err.Error())
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(data)

		} else {
			p := PageFromRequest(r)
			var (
				urls []*core.Url
				err  error
			)
			if fetched, _ := reqParamBool("fetched", r); fetched {
				urls, err = core.FetchedUrls(appDB, p.Size, p.Offset())
			} else {
				urls, err = core.ListUrls(store, p.Size, p.Offset())
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Debug(err.Error())
				return
			}

			data, err := json.MarshalIndent(urls, "", "  ")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Debug(err.Error())
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	default:
		NotFoundHandler(w, r)
	}
}

func CrawlingSourcesHandler(w http.ResponseWriter, r *http.Request) {
	// p := PageFromRequest(r)
	// urls, err := core.CrawlingSources(appDB, p.Size, p.Offset())
	// if err != nil {
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }
	urls := make([]string, len(crawlingUrls))
	for i, u := range crawlingUrls {
		urls[i] = u.String()
	}

	data, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		log.Debug(err.Error())
		return
	}
	w.Write(data)
}
