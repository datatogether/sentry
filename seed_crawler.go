package main

import (
	"github.com/datatogether/core"
	"net/http"
	"time"

	"github.com/PuerkitoBio/fetchbot"
)

var (
	// seedFetcher is a sideband GET-only fetcher
	// to snatch urls that look like files as they cross the main
	// crawler
	seedFetcher *fetchbot.Fetcher
	// que for content GET's
	seedQueue *fetchbot.Queue
	// chan to stop the crawler
	stopSeedCrawler chan bool
)

// startCrawling initializes the crawler, queue, stopCrawler channel, and
// crawlingUrls slice
func startCrawlingSeeds() {
	// Create the muxer
	mux := fetchbot.NewMux()

	// Handle all errors the same
	mux.HandleErrors(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		log.Infof("content res error - %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
		mu.Lock()
		delete(enqued, ctx.Cmd.URL().String())
		mu.Unlock()
	}))

	// Handle GET requests for html responses, to parse the body and enqueue all links as HEAD requests.
	mux.Response().Method("GET").Handler(fetchbot.HandlerFunc(
		func(ctx *fetchbot.Context, res *http.Response, err error) {

			u := &core.Url{Url: ctx.Cmd.URL().String()}
			if err := u.Read(store); err != nil {
				// log.Printf("[ERR] url read error: %s - (%s) - %s\n", ctx.Cmd.URL(), NormalizeURL(ctx.Cmd.URL()), err)
				log.Infof("content url read error: %s - %s\n", u.Url, err)
				return
			}

			mu.Lock()
			delete(enqued, u.Url)
			mu.Unlock()

			_, links, err := u.HandleGetResponse(store, res)
			if err != nil {
				log.Info(err.Error())
				return
			}

			// Enqueue all links as HEAD requests
			if err := enqueueDstLinks(u, links, queue); err != nil {
				log.Info(err.Error())
			}
		}))

	// Create the Fetcher, handle the logging first, then dispatch to the Muxer
	h := logHandler("C", mux)

	seedFetcher = fetchbot.New(h)
	seedFetcher.DisablePoliteness = !cfg.Polite
	seedFetcher.CrawlDelay = time.Duration(cfg.CrawlDelaySeconds) * time.Second

	// Start processing
	log.Info("starting C crawler (seeds)")
	q := seedFetcher.Start()
	seedQueue = q

	stopFunc := q.Close
	stopSeedCrawler = make(chan bool)
	go func() {
		<-stopSeedCrawler
		log.Info("stopping C crawler (seeds)")
		stopFunc()
	}()

	q.Block()
}
