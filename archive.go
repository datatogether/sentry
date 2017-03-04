package main

import (
	"time"
)

// ArchiveUrl GET's a url and if it's an HTML page, any links it directly references
func ArchiveUrl(db sqlQueryExecable, url string, done func(err error)) (*Url, []*Link, error) {
	u := &Url{Url: url}
	if _, err := u.ParsedUrl(); err != nil {
		done(nil)
		return nil, nil, err
	}

	if err := u.Read(db); err != nil {
		if err == ErrNotFound {
			if err := u.Insert(db); err != nil {
				done(nil)
				return nil, nil, err
			}
		} else {
			done(nil)
			return nil, nil, err
		}
	}

	// Perform GET request
	links, err := u.Get(db, func(err error) {
		if err != nil {
			done(err)
		}
	})
	if err != nil {
		done(nil)
		return u, links, err
	}

	tasks := len(links) + 1
	errs := make(chan error, tasks)
	taskDone := func(err error) {
		errs <- err
	}

	go func(db sqlQueryExecable, links []*Link) {
		// GET each destination link from this page in parallel
		for _, l := range links {
			if _, err := l.Dst.Get(db, taskDone); err != nil {
				logger.Println(err.Error())
			}

			// need a sleep here to avoid bombing server with requests
			// tooooo hard
			time.Sleep(cfg.CrawlDelaySeconds)
		}
	}(db, links)

	go func() {
		for i := 0; i < tasks; i++ {
			if err := <-errs; err != nil {
				done(err)
				return
			}
		}
		done(nil)
	}()

	return u, links, err
}
