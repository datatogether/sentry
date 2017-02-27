package main

import (
	"encoding/json"
	"time"
)

// A capture is a record of a GET request to a url
type Capture struct {
	Url      string    `json:"url"`
	Created  time.Time `json:"date"`
	Status   int       `json:"status,omitempty"`
	Duration int64     `json:"downloadTook,omitempty"`
	Hash     string    `json:"hash,omitempty"`
	Headers  []string  `json:"headers,omitempty"`
}

// WriteCapture creates a capture record in the DB
func WriteCapture(db sqlQueryExecable, u *Url, t time.Time) error {
	data, err := json.Marshal(u.Headers)
	if err != nil {
		return err
	}

	_, err = db.Exec("insert into captures values ($1, $2, $3, $4)", u.Url, t.Unix(), u.Status, u.DownloadTook, u.Hash, data)
	return err
}

func CapturesForUrl(db sqlQueryable, url string) ([]*Capture, error) {
	res, err := db.Query("select url, created, status, duration, hash, headers from captures where url = $1", url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	captures := make([]*Capture, 0)
	for res.Next() {
		c := &Capture{}
		if err := c.UnmarshalSQL(res); err != nil {
			return nil, err
		}
		captures = append(captures, c)
	}

	return captures, nil
}

func (c *Capture) UnmarshalSQL(row sqlScannable) error {
	var (
		url, hash         string
		duration, created int64
		status            int
		headerData        []byte
	)

	if err := row.Scan(&url, &created, &status, &duration, &hash, &headerData); err != nil {
		return err
	}

	var headers []string
	if headerData != nil {
		if err := json.Unmarshal(headerData, &headers); err != nil {
			return err
		}
	}

	*c = Capture{
		Url:      url,
		Created:  time.Unix(created, 0),
		Status:   status,
		Duration: duration,
		Hash:     hash,
		Headers:  headers,
	}

	return nil
}
