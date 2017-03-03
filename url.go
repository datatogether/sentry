package main

import (
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

type Url struct {
	Url           string        `json:"url"`
	Created       time.Time     `json:"created"`
	Updated       time.Time     `json:"updated"`
	LastGet       *time.Time    `json:"lastGet,omitempty"`
	LastHead      *time.Time    `json:"lastHead,omitempty"`
	Status        int           `json:"status,omitempty"`
	ContentType   string        `json:"contentType,omitempty"`
	ContentLength int64         `json:"contentLength,omitempty"`
	Title         string        `json:"title,omitempty"`
	Id            string        `json:"id,omitempty"`
	DownloadTook  int           `json:"downloadTook,omitempty"`
	HeadersTook   int           `json:"headersTook,omitempty"`
	Headers       []string      `json:"headers,omitempty"`
	Meta          []interface{} `json:"meta,omitempty"`
	Hash          string        `json:"hash,omitempty"`
}

// ParsedUrl calls url.Parse on the url's string field
func (u *Url) ParsedUrl() (*url.URL, error) {
	return url.Parse(u.Url)
}

// Archive GET's a url and all linked urls
func (u *Url) Archive(db sqlQueryExecable) error {
	if err := u.Read(db); err != nil {
		if err == ErrNotFound {
			if err := u.Insert(db); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	links, err := u.Get(db)
	if err != nil {
		return err
	}

	go func(db sqlQueryExecable, links []*Link) {
		// GET each destination link from this page in parallel
		for _, l := range links {
			go func(db sqlQueryExecable, u *Url) {
				if _, err := u.Get(db); err != nil {
					logger.Println(err.Error())
				}
				// need a sleep here to avoid bombing server with requests
				time.Sleep(cfg.CrawlDelaySeconds)
			}(db, l.Dst)
		}
	}(db, links)

	return err
}

// Issue a GET request to this URL if it's eligible for one
func (u *Url) Get(db sqlQueryExecable) (links []*Link, err error) {
	// TODO - should screen to keep GET's within whitelisted domains
	if !u.ShouldEnqueueGet() {
		// we've fetched this url recently, bail.
		return u.ReadDstLinks(db)
	}

	res, err := http.Get(u.Url)
	if err != nil {
		return nil, err
	}

	return u.processGetResponse(db, res)
}

// processResponse
func (u *Url) processGetResponse(db sqlQueryExecable, res *http.Response) (links []*Link, err error) {
	f, err := NewFileFromRes(u.Url, res)
	if err != nil {
		// logger.Printf("[ERR] generating response file: %s - %s\n", u.Url, err)
		return
	}

	// universally recorded responses:
	u.Status = res.StatusCode
	u.ContentLength = res.ContentLength
	u.ContentType = res.Header.Get("Content-Type")
	u.Headers = rawHeadersSlice(res)
	u.Hash = f.Hash

	now := time.Now()
	u.LastGet = &now

	if u.ShouldPutS3() {
		go func() {
			if err := f.PutS3(); err != nil {
				logger.Printf("[ERR] putting file to S3: %s - %s\n", u.Url, err)
			}
		}()
	}

	go func() {
		if err := WriteSnapshot(db, u); err != nil {
			logger.Println("write url snapshot error:", err.Error())
		}
	}()

	// additional processing for html documents
	if strings.Contains(strings.ToLower(u.ContentType), "text/html") {
		var doc *goquery.Document
		// Process the body to find links
		doc, err = goquery.NewDocumentFromReader(f.Data)
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

	return links, nil
}

func (u *Url) ReadSrcLinks(db sqlQueryable) ([]*Link, error) {
	res, err := db.Query("select urls.url, urls.created, urls.updated, last_get, status, content_type, content_length, title, id, headers_took, download_took, headers, meta, hash from urls, links where links.dst = $1 and links.src = urls.url", u.Url)
	// res, err := db.Query(fmt.Sprintf("select %s from links where src = $1", linkCols()), u.Url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	links := make([]*Link, 0)
	for res.Next() {
		src := &Url{}
		if err := src.UnmarshalSQL(res); err != nil {
			return nil, err
		}
		l := &Link{
			Src: src,
			Dst: u,
		}
		links = append(links, l)
	}

	return links, nil
}

func (u *Url) ReadDstLinks(db sqlQueryable) ([]*Link, error) {
	res, err := db.Query("select urls.url, urls.created, urls.updated, last_get, status, content_type, content_length, title, id, headers_took, download_took, headers, meta, hash from urls, links where links.src = $1 and links.dst = urls.url", u.Url)
	// res, err := db.Query(fmt.Sprintf("select %s from links where src = $1", linkCols()), u.Url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	links := make([]*Link, 0)
	for res.Next() {
		dst := &Url{}
		if err := dst.UnmarshalSQL(res); err != nil {
			return nil, err
		}
		l := &Link{
			Src: u,
			Dst: dst,
		}
		links = append(links, l)
	}

	return links, nil
}

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

func (u *Url) ReadContexts(db sqlQueryable) ([]*UrlContext, error) {
	res, err := db.Query(fmt.Sprintf("select %s from context where context.url = $1", urlContextCols()), u.Url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	contexts := make([]*UrlContext, 0)
	for res.Next() {
		c := &UrlContext{}
		if err := c.UnmarshalSQL(res); err != nil {
			return nil, err
		}
		contexts = append(contexts, c)
	}

	return contexts, nil
}

// isFetchable filters to only usable urls. & schemes
func (u *Url) isFetchable() bool {
	_u, err := u.ParsedUrl()
	if err != nil {
		logger.Println(err.Error())
		return false
	}
	if _u.Scheme == "" || _u.Scheme == "http" || _u.Scheme == "https" {
		return true
	}
	return false
}

// ShouldFetch returns weather the url should be added to the queue for updating
// should return true if the url is new, or if we haven't checked this url in a while
func (u *Url) ShouldEnqueueGet() bool {
	return enqued[u.Url] == "" && u.isFetchable() && (u.LastGet == nil || u.LastGet.IsZero() || time.Since(*u.LastGet) > cfg.StaleDuration())
}

func (u *Url) ShouldEnqueueHead() bool {
	return enqued[u.Url] == "" && u.isFetchable() && (u.LastHead == nil || u.LastHead.IsZero() || time.Since(*u.LastHead) > cfg.StaleDuration())
}

func (u *Url) ShouldPutS3() bool {
	return true
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
	_, err := db.Exec(fmt.Sprintf("insert into urls (%s) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)", urlCols()), u.SQLArgs()...)
	if err != nil {
		logger.Println(err.Error())
		logger.Println(u.SQLArgs())
	}
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
	_, err := db.Exec("update urls set created=$2, updated=$3, last_head=$4, last_get=$5, status=$6, content_type=$7, content_length=$8, title=$9, id=$10, headers_took=$11, download_took=$12, headers=$13, meta=$14, hash=$15 where url = $1", u.SQLArgs()...)
	if err != nil {
		logger.Println(err.Error())
		logger.Println(u.SQLArgs())
	}
	return err
}

// Delete a url, should only do for erronious additions
func (u *Url) Delete(db sqlQueryExecable) error {
	_, err := db.Exec("delete from urls where url = $1", u.Url)
	if err != nil {
		logger.Println(err, u)
	}
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
			logger.Printf("error: resolve URL %s - %s\n", val, err)
			return
		}

		dst := &Url{Url: address.String()}
		// Check to see if url exists, creating if not
		if err = dst.Read(db); err != nil {
			if err == ErrNotFound {
				if err = dst.Insert(db); err != nil {
					logger.Println(err.Error())
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
					logger.Println(err.Error())
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

func urlCols() string {
	return "url, created, updated, last_head, last_get, status, content_type, content_length, title, id, headers_took, download_took, headers, meta, hash"
}

func (u *Url) UnmarshalSQL(row sqlScannable) (err error) {
	var (
		rawurl, mime, title, id, hash string
		created, updated              time.Time
		lastGet, lastHead             *time.Time
		length                        int64
		headersTook, downloadTook     int
		headerBytes, metaBytes        []byte
		status                        int
	)

	if err := row.Scan(&rawurl, &created, &updated, &lastHead, &lastGet, &status, &mime, &length, &title, &id, &headersTook, &downloadTook, &headerBytes, &metaBytes, &hash); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		logger.Println(err.Error())
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

	var meta []interface{}
	if metaBytes != nil {
		meta = []interface{}{}
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

func (u *Url) HeadersMap() (headers map[string]string) {
	headers = map[string]string{}
	for i, s := range u.Headers {
		if i%2 == 0 {
			headers[s] = u.Headers[i+1]
		}
	}
	return
}

func (u *Url) Metadata(db sqlQueryable) (*Meta, error) {
	contexts, err := u.ReadContexts(db)
	if err != nil {
		return nil, err
	}

	ibl, err := u.InboundLinks(db)
	if err != nil {
		return nil, err
	}

	obl, err := u.OutboundLinks(db)
	if err != nil {
		return nil, err
	}

	var sha string
	if len(u.Hash) > 4 {
		sha = u.Hash[3:]
	}

	return &Meta{
		Url:           u.Url,
		Date:          u.LastGet,
		HeadersTook:   u.HeadersTook,
		Id:            u.Id,
		Status:        u.Status,
		RawHeaders:    u.Headers,
		Headers:       u.HeadersMap(),
		DownloadTook:  u.DownloadTook,
		Sha256:        sha,
		Multihash:     u.Hash,
		Contexts:      contexts,
		InboundLinks:  ibl,
		OutboundLinks: obl,
	}, nil
}
