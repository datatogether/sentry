package main

import (
	"time"
	"testing"
	"net/http"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/core"
)

// Define the ./static dir to be the html directory
func testRoutes() *http.ServeMux {
	m := http.NewServeMux()
	m.Handle( "/", http.FileServer( http.Dir( "./static" ) ) )
	return m
}


const(
	MY_PORT = ":8002"
	MY_URL = "http://127.0.0.1" + MY_PORT
)

// Initializes the crawler and a static pages server to look up the URLs that are caught
func TestStartCrawling(t *testing.T) {
	go startCrawling()
	cases := []struct {
		url string
		expected bool //True if it must be in DB
	}{
		{ "http://youShouldNotHaveThis.jingle", false },
		{ "ThisIsNotALink.custom", false },
		{ "http://ThisIsNotALink.customdomain", false },
		{ MY_URL + "chinchila.jpg", false },

		{ MY_URL, true },
		{ MY_URL + "/gallery.html", true },
		{ MY_URL + "/styles.css", true },
		{ "https://google.com/", true },
		{ "http://reddit.com", true },
		{ "ftp://ftp.6te.net/", true },
		{ "mailto:somerandomemai@domain.co.ck", true },
		{ "http://yahoo.com", true },
	}
	s := &http.Server{}
	s.Handler = testRoutes()
	s.Addr = MY_PORT

	// Start the test case one second after the crawler start working.
	ticker := time.NewTicker( 1 * time.Second )
	go func( s *http.Server ) {
		<- ticker.C
		for i, c := range cases {
			store := sql_datastore.NewDatastore( appDB )
			if err := store.Register( &core.Url{} ); err != nil {
				t.Error( err.Error() )
				return
			}

			urls, err := core.ListUrls( store, 10, 0 )
			if err != nil {
				t.Error( err.Error() )
			}

			j := 0
			for _,u := range urls {
				if u.Url == c.url {
					if !c.expected {
						t.Error( "Error on test case", i, "expected", c.expected, "but found in DB" )
					}
					break
				}
				j++
			}
			// If not found in DB and is expected to be true
			if j == len( urls ) && c.expected {
				t.Error( "Error on test case", i, "expected", c.expected, "but not found in DB url:", c.url )
			}
		}

		// End of test case, stops crawlers and shutdown server.
		close( stopContentCrawler )
		close( stopCrawler )
		s.Shutdown( nil )
		ticker.Stop()
	}( s )

	s.ListenAndServe()
}

