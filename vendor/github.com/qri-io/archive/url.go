package archive

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pborman/uuid"
)

// URL represents... a url.
type Url struct {
	// A Url is uniquely identified by URI string without
	// any normalization. Url strings must always be absolute.
	Url string `json:"url"`
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`

	// Timestamp for most recent GET request
	LastGet *time.Time `json:"lastGet,omitempty"`
	// Timestamp for most revent HEAD request
	LastHead *time.Time `json:"lastHead,omitempty"`

	// Returned HTTP status code
	Status int `json:"status,omitempty"`
	// Returned HTTP 'Content-Type' header
	ContentType string `json:"contentType,omitempty"`
	// Result of mime sniffing to GET response body, as detailed at https://mimesniff.spec.whatwg.org
	ContentSniff string `json:"contentSniff,omitempty"`
	// ContentLength in bytes, will be the header value if only a HEAD request has been issued
	// After a valid GET response, it will be set to the length of the returned response
	ContentLength int64 `json:"contentLength,omitempty"`

	// HTML Title tag attribute
	Title string `json:"title,omitempty"`
	// uuid assigend on creation for legacy purposes. Will be depricated in the future.
	// content should be uniquely identified by either url (mutable) or hash (immutable)
	// instead of uuid
	Id string `json:"id,omitempty"`

	// Time remote server took to transfer content in miliseconds.
	// TODO - currently not implemented
	DownloadTook int `json:"downloadTook,omitempty"`
	// Time taken to  in miliseconds. currently not implemented
	HeadersTook int `json:"headersTook,omitempty"`

	// key-value slice of returned headers from most recent HEAD or GET request
	// stored in the form [key,value,key,value...]
	Headers []string `json:"headers,omitempty"`
	// any associative metadata
	Meta map[string]interface{} `json:"meta,omitempty"`

	// Hash is a multihash sha-256 of res.Body
	Hash string `json:"hash,omitempty"`
}

// ParsedUrl is a convenience wrapper around url.Parse
func (u *Url) ParsedUrl() (*url.URL, error) {
	return url.Parse(u.Url)
}

// Issue a GET request to this URL if it's eligible for one
func (u *Url) Get(db sqlQueryExecable, done func(err error)) (links []*Link, err error) {
	// TODO - should screen to keep GET's within whitelisted domains?
	if !u.ShouldEnqueueGet() {
		// we've fetched this url recently, bail with already-stored links
		done(nil)
		return ReadDstLinks(db, u)
	}

	// actual get request using http.DefaultClient
	res, err := http.Get(u.Url)
	if err != nil {
		done(nil)
		return nil, err
	}

	return u.HandleGetResponse(db, res, done)
}

// read headers as a slice of strings in the form [key,value,key,value...] from an http response
func rawHeadersSlice(res *http.Response) (headers []string) {
	for key, val := range res.Header {
		headers = append(headers, []string{key, strings.Join(val, ",")}...)
	}
	return
}

// HandleGetResponse performs all necessary actions in response to a GET request, regardless
// of weather it came from a crawl or archive request
func (u *Url) HandleGetResponse(db sqlQueryExecable, res *http.Response, done func(err error)) (links []*Link, err error) {
	f, err := NewFileFromRes(u.Url, res)
	if err != nil {
		done(err)
		return
	}

	// universally recorded responses:
	u.Status = res.StatusCode
	u.ContentLength = int64(len(f.Data))
	u.ContentType = res.Header.Get("Content-Type")
	u.ContentSniff = http.DetectContentType(f.Data)
	u.Headers = rawHeadersSlice(res)
	u.Hash = f.Hash

	now := time.Now()
	u.LastGet = &now

	tasks := 0
	c := make(chan error, 2)

	go func() {
		tasks++
		c <- WriteSnapshot(db, u)
	}()

	if u.ShouldPutS3() {
		tasks++
		go func() {
			c <- f.PutS3()
		}()
	}

	// additional processing for html documents
	if strings.Contains(strings.ToLower(u.ContentType), "text/html") {
		var doc *goquery.Document
		// Process the body to find links
		doc, err = goquery.NewDocumentFromReader(bytes.NewBuffer(f.Data))
		if err != nil {
			return
		}

		u.Title = doc.Find("title").Text()
		links, err = u.ExtractDocLinks(db, doc)
		if err != nil {
			return
		}
	}

	err = u.Update(db)
	if err != nil {
		return
	}

	go func() {
		for i := 0; i < tasks; i++ {
			err := <-c
			if err != nil {
				done(err)
				return
			}
		}
		done(nil)
	}()

	return links, nil
}

// InboundLinks returns a slice of url strings that link to this url
func (u *Url) InboundLinks(db sqlQueryable) ([]string, error) {
	res, err := db.Query("select src from links where dst = $1", u.Url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	links := make([]string, 0)
	for res.Next() {
		var l string
		if err := res.Scan(&l); err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	return links, nil
}

// Outbound returns a slice of url strings that this url links to
func (u *Url) OutboundLinks(db sqlQueryable) ([]string, error) {
	res, err := db.Query("select dst from links where src = $1", u.Url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	links := make([]string, 0)
	for res.Next() {
		var l string
		if err := res.Scan(&l); err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	return links, nil
}

// ReadContexts reads all context information contributed about this url
// func (u *Url) ReadContexts(db sqlQueryable) ([]*UrlContext, error) {
// 	res, err := db.Query(fmt.Sprintf("select %s from context where context.url = $1", urlContextCols()), u.Url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer res.Close()

// 	contexts := make([]*UrlContext, 0)
// 	for res.Next() {
// 		c := &UrlContext{}
// 		if err := c.UnmarshalSQL(res); err != nil {
// 			return nil, err
// 		}
// 		contexts = append(contexts, c)
// 	}

// 	return contexts, nil
// }

// isFetchable filters to only usable urls & schemes
// this filters out stuff like mailto:// and ftp:// schemes
func (u *Url) isFetchable() bool {
	_u, err := u.ParsedUrl()
	if err != nil {
		return false
	}
	if _u.Scheme == "" || _u.Scheme == "http" || _u.Scheme == "https" {
		return true
	}
	return false
}

// ShouldEnqueueHead returns weather the url can be added to the que for a HEAD request.
// It should return true if:
// * the url is of http / https scheme
// * has never been GET'd or hasn't been GET'd for a period longer than the stale duration
func (u *Url) ShouldEnqueueHead() bool {
	return u.isFetchable() && (u.LastHead == nil || u.LastHead.IsZero() || time.Since(*u.LastHead) > StaleDuration)
}

// ShouldEnqueueGet returns weather the url can be added to the que for a GET request.
// keep in mind only urls who's domain are are marked crawl : true in the domains list
// will be candidates for GET requests.
// It should return true if:
// * the url is of http / https scheme
// * has never been GET'd or hasn't been GET'd for a period longer than the stale duration
func (u *Url) ShouldEnqueueGet() bool {
	return u.isFetchable() && (u.LastGet == nil || u.LastGet.IsZero() || time.Since(*u.LastGet) > StaleDuration)
}

// ShouldPutS3 is a chance to override weather the content should be stored
func (u *Url) ShouldPutS3() bool {
	return true
}

// File leverages a url's hash to generate a file that can have it's bytes read back
func (u *Url) File() (*File, error) {
	if u.Hash == "" {
		return nil, fmt.Errorf("hash required to generate file from url")
	}

	return &File{Hash: u.Hash}, nil
}

// Read url from db
func (u *Url) Read(db sqlQueryable) error {
	if u.Url != "" {
		row := db.QueryRow(fmt.Sprintf("select %s from urls where url = $1", urlCols()), u.Url)
		return u.UnmarshalSQL(row)
	}
	return ErrNotFound
}

// Insert (create)
func (u *Url) Insert(db sqlQueryExecable) error {
	u.Created = time.Now().Round(time.Second)
	u.Updated = u.Created
	u.Id = uuid.New()
	_, err := db.Exec(fmt.Sprintf("insert into urls (%s) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)", urlCols()), u.SQLArgs()...)
	return err
}

// Update url db entry
func (u *Url) Update(db sqlQueryExecable) error {
	u.Updated = time.Now().Round(time.Second)
	if u.ContentLength < -1 {
		u.ContentLength = -1
	}
	if u.Status < -1 {
		u.Status = -1
	}
	_, err := db.Exec("update urls set created=$2, updated=$3, last_head=$4, last_get=$5, status=$6, content_type=$7, content_sniff=$8, content_length=$9, title=$10, id=$11, headers_took=$12, download_took=$13, headers=$14, meta=$15, hash=$16 where url = $1", u.SQLArgs()...)
	return err
}

// Delete a url, should only do for erronious additions
func (u *Url) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from urls where url = $1", u.Url)
	return err
}

// ExtractDocLinks extracts & stores a page's linked documents
// by selecting all a[href] links from a given qoquery document, using
// the receiver *Url as the base
func (u *Url) ExtractDocLinks(db sqlQueryExecable, doc *goquery.Document) ([]*Link, error) {
	pUrl, err := u.ParsedUrl()
	if err != nil {
		return nil, err
	}

	links := make([]*Link, 0)
	// generate a list of normalized links
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr("href")

		// Resolve destination address to source url
		address, err := pUrl.Parse(val)
		if err != nil {
			return
		}

		dst := &Url{Url: address.String()}
		// Check to see if url exists, creating if not
		if err = dst.Read(db); err != nil {
			if err == ErrNotFound {
				if err = dst.Insert(db); err != nil {
					return
				}
			} else {
				return
			}
		}

		// create link
		l := &Link{
			Src: u,
			Dst: dst,
		}

		// confirm link from src to dest exists,
		// creating if not
		if err = l.Read(db); err != nil {
			if err == ErrNotFound {
				if err = l.Insert(db); err != nil {
					return
				}
			} else {
				return
			}
		}

		links = append(links, l)
	})

	return links, nil
}

// HeadersMap formats u.Headers (a string slice) as a map[header]value
func (u *Url) HeadersMap() (headers map[string]string) {
	headers = map[string]string{}
	for i, s := range u.Headers {
		if i%2 == 0 {
			headers[s] = u.Headers[i+1]
		}
	}
	return
}

// Metadata collects up all metadata as
// func (u *Url) Metadata(db sqlQueryable) (*Meta, error) {
// 	contexts, err := u.ReadContexts(db)
// 	if err != nil {
// 		return nil, err
// 	}

// 	ibl, err := u.InboundLinks(db)
// 	if err != nil {
// 		return nil, err
// 	}

// 	obl, err := u.OutboundLinks(db)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var sha string
// 	if len(u.Hash) > 4 {
// 		sha = u.Hash[3:]
// 	}

// 	return &Meta{
// 		Url:           u.Url,
// 		Date:          u.LastGet,
// 		HeadersTook:   u.HeadersTook,
// 		Id:            u.Id,
// 		Status:        u.Status,
// 		ContentSniff:  u.ContentSniff,
// 		RawHeaders:    u.Headers,
// 		Headers:       u.HeadersMap(),
// 		DownloadTook:  u.DownloadTook,
// 		Sha256:        sha,
// 		Multihash:     u.Hash,
// 		Contexts:      contexts,
// 		InboundLinks:  ibl,
// 		OutboundLinks: obl,
// 	}, nil
// }

// standard-form columns for selection from postgres
func urlCols() string {
	return "url, created, updated, last_head, last_get, status, content_type, content_sniff, content_length, title, id, headers_took, download_took, headers, meta, hash"
}

// UnmarshalSQL reads an sql response into the url receiver
// it expects the request to have used urlCols() for selection
func (u *Url) UnmarshalSQL(row sqlScannable) (err error) {
	var (
		rawurl, mime, sniff, title, id, hash string
		created, updated                     time.Time
		lastGet, lastHead                    *time.Time
		length                               int64
		headersTook, downloadTook            int
		headerBytes, metaBytes               []byte
		status                               int
	)

	if err := row.Scan(&rawurl, &created, &updated, &lastHead, &lastGet, &status, &mime, &sniff, &length, &title, &id, &headersTook, &downloadTook, &headerBytes, &metaBytes, &hash); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	var headers []string
	if headerBytes != nil {
		headers = []string{}
		err = json.Unmarshal(headerBytes, &headers)
		if err != nil {
			return err
		}
	}

	var meta map[string]interface{}
	if metaBytes != nil {
		meta = map[string]interface{}{}
		err = json.Unmarshal(metaBytes, &meta)
		if err != nil {
			return err
		}
	}

	if lastGet != nil {
		utc := lastGet.In(time.UTC)
		lastGet = &utc
	}

	if lastHead != nil {
		utc := lastHead.In(time.UTC)
		lastHead = &utc
	}

	*u = Url{
		Created:       created.In(time.UTC),
		Updated:       updated.In(time.UTC),
		LastHead:      lastHead,
		LastGet:       lastGet,
		Url:           rawurl,
		Status:        status,
		ContentType:   mime,
		ContentSniff:  sniff,
		ContentLength: length,
		Title:         title,
		Id:            id,
		HeadersTook:   headersTook,
		DownloadTook:  downloadTook,
		Headers:       headers,
		Meta:          meta,
		Hash:          hash,
	}

	return nil
}

// SQLArgs formats a url struct for inserting / updating into postgres
func (u *Url) SQLArgs() []interface{} {
	headerBytes, err := json.Marshal(u.Headers)
	if err != nil {
		panic(err)
	}
	metaBytes, err := json.Marshal(u.Meta)
	if err != nil {
		panic(err)
	}

	lastGet := u.LastGet
	if lastGet != nil {
		utc := lastGet.In(time.UTC)
		lastGet = &utc
	}

	lastHead := u.LastHead
	if lastHead != nil {
		utc := lastHead.In(time.UTC)
		lastHead = &utc
	}

	return []interface{}{
		u.Url,
		u.Created.In(time.UTC),
		u.Updated.In(time.UTC),
		lastHead,
		lastGet,
		u.Status,
		u.ContentType,
		u.ContentSniff,
		u.ContentLength,
		u.Title,
		u.Id,
		u.HeadersTook,
		u.DownloadTook,
		headerBytes,
		metaBytes,
		u.Hash,
	}
}
