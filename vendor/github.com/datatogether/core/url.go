package core

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/datatogether/ffi"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	"github.com/datatogether/warc"
	"github.com/ipfs/go-datastore"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pborman/uuid"
)

var unwantedMimetypes = map[string]bool{
	"text/html":                 true,
	"text/html; charset=utf-8":  true,
	"text/plain; charset=utf-8": true,
	"text/xml; charset=utf-8":   true,
}

// notContentExtensions is a dictionary of "file extensions" to ignore when
// determining weather or not to prioritize a url for content fetching
var notContentExtensions = map[string]bool{
	".asp":   true,
	".aspx":  true,
	".cfm":   true,
	".html":  true,
	".net":   true,
	".php":   true,
	".xhtml": true,
}

// URL represents... a url.
// TODO - consider renaming to Resource
type Url struct {
	// version 4 uuid
	// urls can/should/must also be be uniquely identified by Url
	Id string `json:"id,omitempty"`
	// A Url is uniquely identified by URI string without
	// any normalization. Url strings must always be absolute.
	Url string `json:"url"`
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created,omitempty"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated,omitempty"`

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

	// best guess at a filename based on url string analysis
	// if you just want to know what type of file this is, this is the field to use.
	FileName string `json:"fileName,omitempty"`

	// HTML Title tag attribute
	Title string `json:"title,omitempty"`

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

	// Url to saved content
	ContentUrl string `json:"contentUrl,omitempty"`

	// Uncrawlable information
	Uncrawlable *Uncrawlable `json:"uncrawlable,omitempty"`
}

func (u Url) DatastoreType() string {
	return "Url"
}

func (u Url) GetId() string {
	return u.Id
}

func (u Url) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s", u.DatastoreType(), u.GetId()))
}

// ParsedUrl is a convenience wrapper around url.Parse
func (u *Url) ParsedUrl() (*url.URL, error) {
	return url.Parse(u.Url)
}

// Issue a GET request to this URL if it's eligible for one
func (u *Url) Get(store datastore.Datastore) (body []byte, links []*Link, err error) {
	// TODO - should screen to keep GET's within whitelisted domains?
	if !u.ShouldEnqueueGet() {
		// we've fetched this url recently, bail with already-stored links
		if sqlds, ok := store.(*sql_datastore.Datastore); ok {
			links, err = ReadDstLinks(sqlds.DB, u)
		}
		return
	}

	// actual get request using http.DefaultClient
	res, err := http.Get(u.Url)
	if err != nil {
		return nil, nil, err
	}

	return u.HandleGetResponse(store, res)
}

// read headers as a slice of strings in the form [key,value,key,value...] from an http response
func rawHeadersSlice(res *http.Response) (headers []string) {
	for key, val := range res.Header {
		headers = append(headers, []string{key, strings.Join(val, ",")}...)
	}
	return
}

func (u *Url) WarcRequest() *warc.Request {
	req := &warc.Request{
		WARCRecordId:  u.Id,
		ContentLength: u.ContentLength,
		WARCTargetURI: u.Url,
	}

	if u.LastGet != nil {
		req.WARCDate = *u.LastGet
	}

	return req
}

// HandleGetResponse performs all necessary actions in response to a GET request, regardless
// of weather it came from a crawl or archive request
func (u *Url) HandleGetResponse(store datastore.Datastore, res *http.Response) (body []byte, links []*Link, err error) {
	var doc *goquery.Document
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	// we're all done with the res reader now.
	res.Body.Close()

	// universally recorded responses:
	u.Status = res.StatusCode
	u.ContentLength = int64(len(body))
	u.ContentType = res.Header.Get("Content-Type")
	u.ContentSniff = http.DetectContentType(body)
	u.Headers = rawHeadersSlice(res)
	// TODO - gotta set this only after adding to IPFS
	// u.Hash = f.Hash

	now := time.Now()
	u.LastGet = &now

	tasks := 0
	c := make(chan error, 2)

	// additional processing for html documents.
	// sometimes xhtml documents can come back as text/plain, thus the text/plain addition
	if u.ContentSniff == "text/html; charset=utf-8" || u.ContentSniff == "text/plain; charset=utf-8" {
		// Process the body to find links
		doc, err = goquery.NewDocumentFromReader(bytes.NewBuffer(body))
		if err != nil {
			return
		}

		u.Title = doc.Find("title").Text()
	} else if !unwantedMimetypes[u.ContentSniff] {
		// handle possible content links
		if filename, err := ffi.FilenameFromUrlString(u.Url); err == nil {
			ext := filepath.Ext(filename)

			// attempt to set file type extenion by checking it against ffi's whitelist of extensions
			_, err := ffi.ExtensionMimeType(ext)

			if !notContentExtensions[ext] && ext != "" && err == nil {
				u.FileName = filename
			} else if err != nil {
				// TODO - should this be reported as an error?
				fmt.Println(err.Error())
			}
		}
	}

	err = u.Save(store)
	if err != nil {
		return
	}

	go func() {
		tasks++
		c <- WriteSnapshot(store, u)
	}()

	if doc != nil {
		go func() {
			tasks++
			links, err = u.ExtractDocLinks(store, doc)
			if err != nil {
				return
			}
		}()
	}

	for i := 0; i < tasks; i++ {
		err = <-c
		if err != nil {
			break
		}
	}

	return
}

// InboundLinks returns a slice of url strings that link to this url
func (u *Url) InboundLinks(db sqlutil.Queryable) ([]string, error) {
	res, err := db.Query(qUrlInboundLinkUrlStrings, u.Url)
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
func (u *Url) OutboundLinks(db sqlutil.Queryable) ([]string, error) {
	res, err := db.Query(qUrlOutboundLinkUrlStrings, u.Url)
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
// func (u *Url) ReadContexts(db sqlutil.Queryable) ([]*UrlContext, error) {
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

// SuspectedContentUrl examines the url string, returns true
// if there's a reasonable chance the url leads to content
func (u *Url) SuspectedContentUrl() bool {
	if unwantedMimetypes[u.ContentSniff] {
		return false
	}

	filename, err := ffi.FilenameFromUrlString(u.Url)
	if err != nil {
		return false
	}

	ext := filepath.Ext(filename)
	if filename == "" || notContentExtensions[ext] || ext == "." || ext == "" {
		return false
	}

	return true
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
func (u *Url) Read(store datastore.Datastore) error {
	if u.Id != "" {
		ci, err := store.Get(u.Key())
		if err != nil {
			return err
		}

		got, ok := ci.(*Url)
		if !ok {
			return ErrInvalidResponse
		}
		*u = *got
		return nil
	} else {
		// TODO - figure out a way to query stores by url...
		if sqlstore, ok := store.(*sql_datastore.Datastore); ok {
			if u.Url != "" {
				row := sqlstore.DB.QueryRow(qUrlByUrlString, u.Url)
				return u.UnmarshalSQL(row)
			} else if u.Hash != "" {
				row := sqlstore.DB.QueryRow(qUrlByHash, u.Hash)
				return u.UnmarshalSQL(row)
			}
		}
	}
	return ErrNotFound
}

func (u *Url) Save(store datastore.Datastore) (err error) {
	var exists bool

	if u.Id != "" {
		exists, err = store.Has(u.Key())
		if err != nil {
			return err
		}
	} else if sqls, ok := store.(*sql_datastore.Datastore); ok {
		// if no Id is set, attempt to set one
		if u.Url != "" {
			row := sqls.DB.QueryRow(qUrlByUrlString, u.Url)
			prev := &Url{}
			if err := prev.UnmarshalSQL(row); err == nil {
				u.Id = prev.Id
				exists = true
			}
		}
	}

	// TODO - support fetching ID via url entry
	// 	// Need to fetch ID
	// 	if u.Url != "" && u.Id == "" {
	// 		prev := &Url{Url: u.Url}
	// 		if err := prev.Read(store); err != ErrNotFound {
	// 			return err
	// 		}
	// 		u.Id = prev.Id
	// 	}

	if err = u.validate(); err != nil {
		return
	}

	if !exists {
		u.Id = uuid.New()
		u.Created = time.Now().Round(time.Second).In(time.UTC)
		u.Updated = u.Created
	} else {
		u.Updated = time.Now().Round(time.Second).In(time.UTC)
	}

	return store.Put(u.Key(), u)
}

func (u *Url) validate() error {
	if u.ContentLength < -1 {
		u.ContentLength = -1
	}
	if u.Status < -1 {
		u.Status = -1
	}
	return nil
}

// Delete a url, should only do for erronious additions
func (u *Url) Delete(store datastore.Datastore) error {
	return store.Delete(u.Key())
}

// ExtractDocLinks extracts & stores a page's linked documents
// by selecting all a[href] links from a given qoquery document, using
// the receiver *Url as the base
func (u *Url) ExtractDocLinks(store datastore.Datastore, doc *goquery.Document) ([]*Link, error) {
	pUrl, err := u.ParsedUrl()
	if err != nil {
		return nil, err
	}

	links := make([]*Link, 0)
	// generate a list of normalized links
	doc.Find("[href]").Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr("href")

		// Resolve destination address to source url
		address, err := pUrl.Parse(val)
		if err != nil {
			return
		}

		dst := &Url{Url: address.String()}
		// Check to see if url exists, creating if not
		if err = dst.Read(store); err != nil {
			if err == ErrNotFound {
				if err = dst.Save(store); err != nil {
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
		if err = l.Read(store); err != nil {
			if err == ErrNotFound {
				if err = l.Insert(store); err != nil {
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

func (u *Url) NewSQLModel(key datastore.Key) sql_datastore.Model {
	return &Url{
		Id:   key.Name(),
		Url:  u.Url,
		Hash: u.Hash,
	}
}

func (u *Url) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qUrlsCreateTable
	case sql_datastore.CmdSelectOne:
		if u.Id != "" {
			return qUrlById
		} else if u.Hash != "" {
			return qUrlByHash
		} else {
			return qUrlByUrlString
		}
	case sql_datastore.CmdExistsOne:
		if u.Id != "" {
			return qUrlExistsById
		} else if u.Hash != "" {
			return qUrlExistsByHash
		} else {
			return qUrlExistsByUrlString
		}
	case sql_datastore.CmdInsertOne:
		return qUrlInsert
	case sql_datastore.CmdUpdateOne:
		return qUrlUpdate
	case sql_datastore.CmdDeleteOne:
		return qUrlDelete
	case sql_datastore.CmdList:
		return qUrlsList
	default:
		return ""
	}
}

// SQLArgs formats a url struct for inserting / updating into postgres
func (u *Url) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdList:
		return []interface{}{}
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne:
		// fmt.Println(u)
		if u.Id != "" {
			return []interface{}{u.Id}
		} else if u.Hash != "" {
			return []interface{}{u.Hash}
		} else {
			return []interface{}{u.Url}
		}
	case sql_datastore.CmdDeleteOne:
		return []interface{}{u.Url}
	default:
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
			u.FileName,
			u.Title,
			u.Id,
			u.HeadersTook,
			u.DownloadTook,
			headerBytes,
			metaBytes,
			u.Hash,
		}
	}
}

// UnmarshalSQL reads an sql response into the url receiver
// it expects the request to have used urlCols() for selection
func (u *Url) UnmarshalSQL(row sqlutil.Scannable) (err error) {
	var (
		rawurl, mime, sniff, title, id, hash, fn string
		created, updated                         time.Time
		lastGet, lastHead                        *time.Time
		length                                   int64
		headersTook, downloadTook                int
		headerBytes, metaBytes                   []byte
		status                                   int
	)

	if err := row.Scan(&rawurl, &created, &updated, &lastHead, &lastGet, &status, &mime, &sniff, &length, &fn, &title, &id, &headersTook, &downloadTook, &headerBytes, &metaBytes, &hash); err != nil {
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
		FileName:      fn,
		Title:         title,
		Id:            id,
		HeadersTook:   headersTook,
		DownloadTook:  downloadTook,
		Headers:       headers,
		Meta:          meta,
		Hash:          hash,
	}

	if u.Hash != "" && u.FileName != "" {
		u.ContentUrl = FileUrl(u)
	}

	return nil
}
