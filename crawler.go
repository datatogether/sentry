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
	"github.com/PuerkitoBio/goquery"
)

var (
	// the fetcher that's doing the crawling
	f *fetchbot.Fetcher

	// Protect access to crawling domains map
	mu sync.Mutex
	// cached list of domains currently crawling
	crawlingDomains = map[string]bool{}
)

func startCrawling() {
	// Create the muxer
	mux := fetchbot.NewMux()

	// Handle all errors the same
	mux.HandleErrors(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		fmt.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
	}))

	// Handle GET requests for html responses, to parse the body and enqueue all links as HEAD requests.
	mux.Response().Method("GET").ContentType("text/html").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {
			addr := NormalizeURL(ctx.Cmd.URL())
			u := &Url{
				Url:  addr,
				Host: addr.Host,
			}

			if err := u.Read(appDB); err != nil {
				fmt.Printf("[ERR] url read error: %s - (%s) - %s\n", ctx.Cmd.URL(), NormalizeURL(ctx.Cmd.URL()), err)
				return
			}

			// if err := u.ReadDomain(appDB); err != nil {
			// 	fmt.Println("[ERR] reading domain for url: %s - %s", ctx.Cmd.URL().Host, err)
			// 	return
			// }

			// Process the body to find links
			doc, err := goquery.NewDocumentFromResponse(res)
			if err != nil {
				fmt.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				return
			}

			u.Title = doc.Find("title").Text()
			u.Status = res.StatusCode
			u.ContentLength = res.ContentLength
			u.ContentType = res.Header.Get("Content-Type")
			links, err := u.DocLinks(doc)
			if err != nil {
				fmt.Printf("[ERR] finding doc links: %s - %s\n", u.Url.String(), err)
				return
			}

			if err := u.Update(appDB); err != nil {
				fmt.Println("[ERR] updating url: %s - %s", u.Url.String(), err)
				return
			}

			// Enqueue all links as HEAD requests
			if err := enqueueDstLinks(appDB, links, ctx); err != nil {
				fmt.Println(err.Error())
			}
		}))

	// Handle HEAD requests for html responses coming from the source host - we don't want
	// to crawl links from other hosts.
	mux.Response().Method("HEAD").ContentType("text/html").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {
			// Normalize the host right out the gate
			addr := NormalizeURL(ctx.Cmd.URL())

			u := &Url{
				Url: addr,
			}

			// if we're currently crawling this url's domain, attept to add it to the
			// queue
			if crawlingDomains[addr.Host] {
				if err := enqueueDomainGet(u, ctx); err != nil {
					fmt.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				}
			} else {
				// we're not crawling this domain, let's save the head info
				if err := u.Read(appDB); err != nil {
					fmt.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				}
				u.Status = res.StatusCode
				u.ContentLength = res.ContentLength
				u.ContentType = res.Header.Get("Content-Type")
				if err := u.Update(appDB); err != nil {
					fmt.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
				}
			}
		}))

	//
	// mux.Response().Method("HEAD").Handler(h)

	// Create the Fetcher, handle the logging first, then dispatch to the Muxer
	h := logHandler(mux)
	// if *stopAtURL != "" || *cancelAtURL != "" {
	// 	stopURL := *stopAtURL
	// 	if *cancelAtURL != "" {
	// 		stopURL = *cancelAtURL
	// 	}
	// 	h = stopHandler(stopURL, *cancelAtURL != "", logHandler(mux))
	// }

	fmt.Println("startin' crawlin'")
	f = fetchbot.New(h)
	f.DisablePoliteness = true
	f.CrawlDelay = 5 * time.Second

	// First mem stat print must be right after creating the fetchbot
	// if *memStats > 0 {
	// 	// Print starting stats
	// 	printMemStats(nil)
	// 	// Run at regular intervals
	// 	runMemStats(f, *memStats)
	// 	// On exit, print ending stats after a GC
	// 	defer func() {
	// 		runtime.GC()
	// 		printMemStats(nil)
	// 	}()
	// }

	// Start processing
	q := f.Start()

	// if a stop or cancel is requested after some duration, launch the goroutine
	// that will stop or cancel.
	// if *stopAfter > 0 || *cancelAfter > 0 {
	// 	after := *stopAfter
	// 	stopFunc := q.Close
	// 	if *cancelAfter != 0 {
	// 		after = *cancelAfter
	// 		stopFunc = q.Cancel
	// 	}

	// 	go func() {
	// 		c := time.After(after)
	// 		<-c
	// 		stopFunc()
	// 	}()
	// }

	// do an initial domain seed
	seedDomains(appDB, q)
	// every half stale-duration, check to see if top levels need to be re-crawled for staleness
	go func() {
		c := time.After(time.Duration(cfg.StaleDuration / 2))
		<-c
		seedDomains(appDB, q)
	}()

	q.Block()
}

func seedDomains(db sqlQueryExecable, q *fetchbot.Queue) error {
	rows, err := db.Query(fmt.Sprintf("select %s from domains where crawl = true", domainCols()))
	if err != nil {
		fmt.Println(err)
		return err
	}

	mu.Lock()
	for rows.Next() {
		d := &Domain{}
		if err := d.UnmarshalSQL(rows); err != nil {
			return err
		}
		crawlingDomains[d.Host] = true
		u, err := d.Url(db)
		if err != nil {
			fmt.Println(err)
			return err
		}

		// only seed domains if we haven't looked at them since their stale date
		fmt.Println("should enqueue", u.Url.String(), "?")
		if u.ShouldEnqueue() {
			fmt.Println("yes")
			_, err := q.SendStringGet(u.Url.String())
			if err != nil {
				return err
			}
		}
	}
	mu.Unlock()
	return nil
}

func enqueueDomainGet(u *Url, ctx *fetchbot.Context) error {
	if u.ShouldEnqueue() {
		_, err := ctx.Q.SendStringGet(u.Url.String())
		return err
	}
	return nil
}

func enqueueDstLinks(db sqlQueryExecable, links []*Link, ctx *fetchbot.Context) error {
	for _, l := range links {
		// Check to see if url exists, creating if not
		if err := l.Dst.Read(db); err != nil {
			if err == ErrNotFound {
				if err := l.Dst.Insert(db); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// confirm link from src to dest exists,
		// creating if not
		if err := l.Read(db); err != nil {
			if err == ErrNotFound {
				if err := l.Insert(db); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		if l.Dst.ShouldEnqueue() {
			if _, err := ctx.Q.SendStringHead(l.Dst.Url.String()); err != nil {
				fmt.Printf("error: enqueue head %s - %s\n", l.Dst.Url.String(), err)
			} else {
				// at this point the destination has been added for a HEAD request.
				// dup[u.String()] = true
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

// logHandler prints the fetch information and dispatches the call to the wrapped Handler.
func logHandler(wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if err == nil {
			fmt.Printf("[%d] %s %s - %s\n", res.StatusCode, ctx.Cmd.Method(), ctx.Cmd.URL(), res.Header.Get("Content-Type"))
		}
		wrapped.Handle(ctx, res, err)
	})
}

func runMemStats(f *fetchbot.Fetcher, tick time.Duration) {
	var mu sync.Mutex
	var di *fetchbot.DebugInfo

	// Start goroutine to collect fetchbot debug info
	go func() {
		for v := range f.Debug() {
			mu.Lock()
			di = v
			mu.Unlock()
		}
	}()
	// Start ticker goroutine to print mem stats at regular intervals
	go func() {
		c := time.Tick(tick)
		for _ = range c {
			mu.Lock()
			printMemStats(di)
			mu.Unlock()
		}
	}()
}

func printMemStats(di *fetchbot.DebugInfo) {
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
	fmt.Println(buf.String())
}
