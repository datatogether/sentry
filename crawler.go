package main

import (
	"bytes"
	"fmt"
	"net/http"
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
	// map of domains currently crawling
	crawlingUrls = map[string]bool{}
	// enqued map of url : method (HEAD|GET) to prevent double-adding
	// to the que
	enqued = map[string]string{}
	// chan to stop the crawler
	stopCrawler chan bool
)

// startCrawling initializes the crawler, queue, stopCrawler channel, and
// crawlingUrls map
func startCrawling() {
	// Create the muxer
	mux := fetchbot.NewMux()

	// Handle all errors the same
	mux.HandleErrors(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		mu.Lock()
		delete(enqued, ctx.Cmd.URL().String())
		mu.Unlock()
		logger.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
	}))

	// Handle GET requests for html responses, to parse the body and enqueue all links as HEAD requests.
	mux.Response().Method("GET").ContentType("text/html").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {

			u := &Url{Url: ctx.Cmd.URL().String()}
			if err := u.Read(appDB); err != nil {
				// logger.Printf("[ERR] url read error: %s - (%s) - %s\n", ctx.Cmd.URL(), NormalizeURL(ctx.Cmd.URL()), err)
				logger.Printf("[ERR] url read error: %s - %s\n", u.Url, err)
				return
			}

			mu.Lock()
			delete(enqued, u.Url)
			mu.Unlock()

			done := func(err error) {
				if err != nil {
					logger.Println(err.Error())
				}
			}

			links, err := u.handleGetResponse(appDB, res, done)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			// Enqueue all links as HEAD requests
			if err := enqueueDstLinks(links, ctx); err != nil {
				fmt.Println(err.Error())
			}
		}))

	// Handle HEAD requests for html responses coming from the source host - we don't want
	// to crawl links from other hosts.
	mux.Response().Method("HEAD").ContentType("text/html").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {
			addr := ctx.Cmd.URL()
			u := &Url{Url: addr.String()}

			mu.Lock()
			enqued[u.Url] = ""
			mu.Unlock()

			if err := u.Read(appDB); err != nil {
				logger.Println("[ERR] %s %s reading - ", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				return
			}

			u.Status = res.StatusCode
			u.ContentLength = res.ContentLength
			u.ContentType = res.Header.Get("Content-Type")
			u.Headers = rawHeadersSlice(res)
			// TODO u.HeadersTook = 0
			now := time.Now()
			u.LastHead = &now

			if err := u.Update(appDB); err != nil {
				logger.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				logger.Printf("%#v", u)
			}

			// if we're currently crawling this url's domain, attept to add it to the
			// queue
			if crawlingUrls[addr.Host] {
				if err := enqueueDomainGet(u, ctx); err != nil {
					logger.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				}
			}
		}))

	// Create the Fetcher, handle the logging first, then dispatch to the Muxer
	h := logHandler(mux)

	logger.Println("starting crawl")
	f = fetchbot.New(h)
	f.DisablePoliteness = !cfg.Polite
	f.CrawlDelay = cfg.CrawlDelaySeconds * time.Second

	// Start processing
	q := f.Start()
	queue = q

	stopFunc := q.Close
	stopCrawler = make(chan bool)
	go func() {
		<-stopCrawler
		stopFunc()
	}()

	// do an initial domain seed
	seedCrawlingUrls(appDB, q)
	seedUrls(appDB, q)

	// every half stale-duration, check to see if top levels need to be re-crawled for staleness
	go func() {
		c := time.After(time.Duration(cfg.StaleDuration() / 2))
		<-c
		seedCrawlingUrls(appDB, q)
		seedUrls(appDB, q)
	}()

	q.Block()
}

func seedCrawlingUrls(db sqlQueryExecable, q *fetchbot.Queue) error {
	urls, err := CrawlingUrls(db)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	for _, c := range urls {
		crawlingUrls[c.Url] = true
		logger.Println("crawling url:", c.Url)

		u, err := c.AsUrl(db)
		if err != nil {
			fmt.Println(err)
			return err
		}
		enqued[u.Url] = "GET"
		_, err = q.SendStringGet(u.Url)
		if err != nil {
			return err
		}
	}

	return nil
}

// try to read a list of unfetched known urls
func seedUrls(db sqlQueryExecable, q *fetchbot.Queue) error {
	mu.Lock()
	defer mu.Unlock()

	logger.Println("checking for unfeched urls:")
	if ufd, err := UnfetchedUrls(db, 100); err == nil && len(ufd) >= 0 {
		logger.Println("adding %d unfetched urls to que", len(ufd))
		for _, u := range ufd {
			_, err = q.SendStringGet(u.Url)
			if err != nil {
				return err
			}
			enqued[u.Url] = "GET"
		}
	}
	return nil
}

func enqueueDomainGet(u *Url, ctx *fetchbot.Context) error {
	// logger.Printf("url: %s, should head: %t, isFetchable: %t", u.Url, u.ShouldEnqueueHead(), u.isFetchable())
	if u.ShouldEnqueueGet() {
		_, err := ctx.Q.SendStringGet(u.Url)
		if err == nil {
			mu.Lock()
			enqued[u.Url] = "GET"
			mu.Unlock()
		}
		return err
	}
	return nil
}

func enqueueDstLinks(links []*Link, ctx *fetchbot.Context) error {
	mu.Lock()
	defer mu.Unlock()

	for _, l := range links {
		// logger.Printf("url: %s, should head: %t, isFetchable: %t", l.Dst.Url, l.Dst.ShouldEnqueueHead(), l.Dst.isFetchable())
		if l.Dst.ShouldEnqueueHead() {
			if _, err := ctx.Q.SendStringHead(l.Dst.Url); err != nil {
				fmt.Printf("error: enqueue head %s - %s\n", l.Dst.Url, err)
			} else {
				// at this point the destination has been added for a HEAD request.
				// dup[u.String()] = true
				enqued[l.Dst.Url] = "HEAD"
			}
		}
	}
	return nil
}

// stopHandler stops the fetcher if the stopurl is reached. Otherwise it dispatches
// the call to the wrapped Handler.
func stopHandler(stopurl string, cancel bool, wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if ctx.Cmd.URL().String() == stopurl {
			fmt.Printf(">>>>> STOP URL %s\n", ctx.Cmd.URL())
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

func rawHeadersSlice(res *http.Response) (headers []string) {
	for key, val := range res.Header {
		headers = append(headers, []string{key, strings.Join(val, ",")}...)
	}
	return
}

// logHandler prints the fetch information and dispatches the call to the wrapped Handler.
func logHandler(wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if err == nil {
			fmt.Printf("[%d] %s %s - %s\n", res.StatusCode, ctx.Cmd.Method(), ctx.Cmd.URL(), res.Header.Get("Content-Type"))
		}
		wrapped.Handle(ctx, res, err)
	})
}

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
