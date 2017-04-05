package main

import (
	"fmt"
	"github.com/qri-io/archive"
	"net/http"
	"time"

	"github.com/PuerkitoBio/fetchbot"
)

var (
	// contentFetcher is a sideband GET-only fetcher
	// to snatch urls that look like files as they cross the main
	// crawler
	contentFetcher *fetchbot.Fetcher
	// que for content GET's
	contentQueue *fetchbot.Queue
	// chan to stop the crawler
	stopContentCrawler chan bool
)

// startCrawling initializes the crawler, queue, stopCrawler channel, and
// crawlingUrls slice
func startCrawlingContent() {
	// Create the muxer
	mux := fetchbot.NewMux()

	// Handle all errors the same
	mux.HandleErrors(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		mu.Lock()
		delete(enqued, ctx.Cmd.URL().String())
		mu.Unlock()
		logger.Printf("[ERR] content %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
	}))

	// Handle GET requests for html responses, to parse the body and enqueue all links as HEAD requests.
	mux.Response().Method("GET").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {

			u := &archive.Url{Url: ctx.Cmd.URL().String()}
			if err := u.Read(appDB); err != nil {
				// logger.Printf("[ERR] url read error: %s - (%s) - %s\n", ctx.Cmd.URL(), NormalizeURL(ctx.Cmd.URL()), err)
				logger.Printf("[ERR] content url read error: %s - %s\n", u.Url, err)
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

			links, err := u.HandleGetResponse(appDB, res, done)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			// Enqueue all links as HEAD requests
			if err := enqueueDstLinks(links, ctx); err != nil {
				fmt.Println(err.Error())
			}
		}))

	// Create the Fetcher, handle the logging first, then dispatch to the Muxer
	h := logHandler(mux)

	contentFetcher = fetchbot.New(h)
	contentFetcher.DisablePoliteness = !cfg.Polite
	contentFetcher.CrawlDelay = cfg.CrawlDelaySeconds * time.Second

	// Start processing
	q := contentFetcher.Start()
	contentQueue = q

	stopFunc := q.Close
	stopContentCrawler = make(chan bool)
	go func() {
		<-stopContentCrawler
		stopFunc()
	}()

	q.Block()
}
