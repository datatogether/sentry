package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/datatogether/core"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/fetchbot"
)

var (
	// the fetcher that's doing the crawling
	f *fetchbot.Fetcher
	// the queue
	queue *fetchbot.Queue
	// Protect access to crawling domains map
	mu sync.Mutex
	// slice of urls currently crawling
	crawlingUrls []*url.URL
	// enqued map of url : method (HEAD|GET) to prevent double-adding
	// to the que
	enqued = map[string]string{}
	// chan to stop the crawler
	stopCrawler chan bool
)

// startCrawling initializes the crawler, queue, stopCrawler channel, and
// crawlingUrls slice
func startCrawling() {
	go startCrawlingContent()

	// Create the muxer
	mux := fetchbot.NewMux()

	// Handle all errors the same
	mux.HandleErrors(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		mu.Lock()
		delete(enqued, ctx.Cmd.URL().String())
		mu.Unlock()

		log.Infof("res error - %s %s - %s", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
	}))

	// Handle GET requests for html responses, to parse the body and enqueue all links as HEAD requests.
	mux.Response().Method("GET").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {

			u := &core.Url{Url: ctx.Cmd.URL().String()}
			if err := u.Read(store); err != nil {
				// log.Infof("[ERR] url read error: %s - (%s) - %s\n", ctx.Cmd.URL(), NormalizeURL(ctx.Cmd.URL()), err)
				log.Infof("url read error: %s - %s", u.Url, err)
				return
			}

			mu.Lock()
			delete(enqued, u.Url)
			mu.Unlock()

			_, links, err := u.HandleGetResponse(store, res)
			if err != nil {
				log.Debugf("error handling get response: %s - %s", ctx.Cmd.URL().String(), err.Error())
				return
			}

			if err := enqueueDstLinks(u, links, ctx.Q); err != nil {
				log.Debugf("enque links error: %s", err.Error())
			}
		}))

	// Handle HEAD requests for html responses coming from the source host - we don't want
	// to crawl links from other hosts.
	mux.Response().Method("HEAD").ContentType("text/html").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {
			addr := ctx.Cmd.URL()

			u := &core.Url{Url: addr.String()}

			mu.Lock()
			enqued[u.Url] = ""
			mu.Unlock()

			if err := u.Read(store); err != nil {
				log.Info("%s %s reading - ", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				return
			}

			u.Status = res.StatusCode
			u.ContentLength = res.ContentLength
			u.ContentType = res.Header.Get("Content-Type")
			u.Headers = rawHeadersSlice(res)
			// TODO u.HeadersTook = 0
			now := time.Now()
			u.LastHead = &now

			if err := u.Save(store); err != nil {
				log.Infof("update error: %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				log.Infof("%#v", u)
			}

			// if we're currently crawling this url's domain, attept to add it to the
			// queue
			if urlIsWhitelisted(addr) {
				if err := enqueueDomainGet(u, ctx); err != nil {
					log.Infof("error enquing domain get: %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				}
			} else {
				log.Debugf("%s isn't whitelisted", addr.String())
			}
		}))

	// Create the Fetcher, handle the logging first, then dispatch to the Muxer
	h := logHandler("A", mux)

	log.Info("starting A crawler (main)")
	f = fetchbot.New(h)
	f.DisablePoliteness = !cfg.Polite
	f.CrawlDelay = time.Duration(cfg.CrawlDelaySeconds) * time.Second

	// Start processing
	q := f.Start()
	queue = q

	stopFunc := q.Close
	stopCrawler = make(chan bool)
	go func() {
		<-stopCrawler
		log.Info("stopping A crawler (main)")
		stopFunc()
	}()

	// do an initial domain seed
	seedCrawlingSources(appDB, q)
	seedUrls(appDB, q, 10)

	// check to see if top levels need to be re-crawled for staleness
	go func() {
		c := time.Tick(time.Minute * 30)
		select {
		case <-c:
			if len(enqued) < 100 {
				log.Info("que is low, adding urls")
				seedCrawlingSources(appDB, q)
				seedUrls(appDB, q, 400)
			}
		}
	}()

	q.Block()
}

// seedCrawlingSources grabs a list of sources that are currently set to crawl
// and adds them to the que
func seedCrawlingSources(db *sql.DB, q *fetchbot.Queue) error {
	urls, err := core.CrawlingSources(db, 200, 0)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	crawlingUrls = make([]*url.URL, len(urls))
	for i, c := range urls {

		log.Debugf("crawling url: %s", c.Url)

		u, err := c.AsUrl(db)
		if err != nil {
			log.Info(err.Error())
			return err
		}

		url, err := u.ParsedUrl()
		if err != nil {
			return err
		}

		crawlingUrls[i] = url
		enqued[u.Url] = "GET"
		_, err = q.SendStringGet(u.Url)
		if err != nil {
			log.Info("error enquing string get", err.Error())
			return err
		}
	}

	return nil
}

// urlIsWhitelisted scans the slice of crawlingUrls to see if we should GET
// the passed-in url
func urlIsWhitelisted(u *url.URL) bool {
	for _, c := range crawlingUrls {
		// TODO - do we need more than host comparison here
		// to avoid crawling all of a site?
		if c.Host == u.Host {
			return true
		}
	}
	return false
}

// try to read a list of unfetched known urls
func seedUrls(db *sql.DB, q *fetchbot.Queue, count int) error {
	mu.Lock()
	defer mu.Unlock()

	if ufd, err := core.UnfetchedUrls(db, count, 0); err == nil && len(ufd) >= 0 {
		i := 0
		for _, unfetched := range ufd {
			u, err := unfetched.ParsedUrl()
			if err != nil {
				return err
			}
			if urlIsWhitelisted(u) {
				_, err = q.SendStringGet(unfetched.Url)
				if err != nil {
					return err
				}
				enqued[unfetched.Url] = "GET"
				i++
			}
		}
		log.Infof("adding %d unfetched urls to que", i)
	}
	return nil
}

// enqueDomainGet adds a url GET request to the que if the url is valid
// for queing & not already enqued
func enqueueDomainGet(u *core.Url, ctx *fetchbot.Context) error {
	// log.Infof("url: %s, should head: %t, isFetchable: %t", u.Url, u.ShouldEnqueueHead(), u.isFetchable())
	if enqued[u.Url] == "" && u.ShouldEnqueueGet() {
		_, err := ctx.Q.SendStringGet(u.Url)
		if err == nil {
			mu.Lock()
			defer mu.Unlock()
			enqued[u.Url] = "GET"
		}
		return err
	} else if enqued[u.Url] == "" {
		log.Debugf("skipped url: %s last head: %s, last get: %s, content type: %s, content sniff: %s", u.Url, u.LastHead, u.LastGet, u.ContentType, u.ContentSniff)
	}
	return nil
}

// enqueDstLinks works through all linked urls
func enqueueDstLinks(u *core.Url, links []*core.Link, q *fetchbot.Queue) error {
	if links == nil || len(links) == 0 {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	heads := 0
	gets := 0
	for _, l := range links {
		// log.Infof("url: %s, should head: %t, isFetchable: %t", l.Dst.Url, l.Dst.ShouldEnqueueHead(), l.Dst.isFetchable())
		if enqued[l.Dst.Url] == "" && l.Dst.ShouldEnqueueHead() {
			// skip the que & go straight to content archiving if it's a
			if l.Dst.SuspectedContentUrl() && contentQueue != nil {
				gets++
				enqued[l.Dst.Url] = "GET"
				contentQueue.SendStringGet(l.Dst.Url)
				continue
			}

			if q != nil {
				if _, err := q.SendStringHead(l.Dst.Url); err != nil {
					log.Debugf("error: enqueue head %s - %s\n", l.Dst.Url, err)
				} else {
					heads++
					enqued[l.Dst.Url] = "HEAD"
				}
			}
		} else {
			if enqued[l.Dst.Url] == "" {
				log.Debugf("skipped url: %s last head: %s, last get: %s", l.Dst.Url, l.Dst.LastHead, l.Dst.LastGet)
			}
		}
	}
	log.Debugf("enqued %d GET, %d HEAD from %d links for source: %s", gets, heads, len(links), u.Url)
	return nil
}

// stopHandler stops the fetcher if the stopurl is reached. Otherwise it dispatches
// the call to the wrapped Handler.
func stopHandler(stopurl string, cancel bool, wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if ctx.Cmd.URL().String() == stopurl {
			log.Infof(">>>>> STOP URL %s\n", ctx.Cmd.URL())
			// generally not a good idea to stop/block from a handler goroutine
			// so do it in a separate goroutine
			go func() {
				if cancel {
					ctx.Q.Cancel()
				} else {
					ctx.Q.Close()
				}
			}()
			return
		}
		wrapped.Handle(ctx, res, err)
	})
}

// construct a slice of [key,val,key,val,...] listing all response headers
func rawHeadersSlice(res *http.Response) (headers []string) {
	for key, val := range res.Header {
		headers = append(headers, []string{key, strings.Join(val, ",")}...)
	}
	return
}

// logHandler prints the fetch information and dispatches the call to the wrapped Handler.
func logHandler(crawlerId string, wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if err == nil {
			log.Infof("[%d] %s %s %s - %s", res.StatusCode, ctx.Cmd.Method(), crawlerId, ctx.Cmd.URL(), res.Header.Get("Content-Type"))
		}
		wrapped.Handle(ctx, res, err)
	})
}

// memStats prints off this server's current memory statistics
func memStats(di *fetchbot.DebugInfo) []byte {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	buf := bytes.NewBuffer(nil)
	buf.WriteString(strings.Repeat("=", 72) + "\n")
	buf.WriteString("Memory Profile:\n")
	buf.WriteString(fmt.Sprintf("\tAlloc: %d Kb\n", mem.Alloc/1024))
	buf.WriteString(fmt.Sprintf("\tTotalAlloc: %d Kb\n", mem.TotalAlloc/1024))
	buf.WriteString(fmt.Sprintf("\tNumGC: %d\n", mem.NumGC))
	buf.WriteString(fmt.Sprintf("\tGoroutines: %d\n", runtime.NumGoroutine()))
	if di != nil {
		buf.WriteString(fmt.Sprintf("\tNumHosts: %d\n", di.NumHosts))
	}
	buf.WriteString(strings.Repeat("=", 72))
	return buf.Bytes()
}

// enquedUrls lists out all urls currently in the que in no particular order
// TODO - order based on position within the que, or at least confirm that
// the order is in fact random
func enquedUrls() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("Enqued Urls:\n")
	i := 1
	for u, v := range enqued {
		if v != "" {
			buf.WriteString(fmt.Sprintf("%d - %s - %s\n", i, v, u))
			i++
		}
	}
	return buf.Bytes()
}
