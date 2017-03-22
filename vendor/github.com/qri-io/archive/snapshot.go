package archive

import (
	"encoding/json"
	"time"
)

// A snapshot is a record of a GET request to a url
// There can be many snapshots of a given url
type Snapshot struct {
	// The url that was requested
	Url string `json:"url"`
	// Time this request was issued
	Created time.Time `json:"date"`
	// Returned Status
	Status int `json:"status,omitempty"`
	// Time to complete response in milliseconds
	Duration int64 `json:"downloadTook,omitempty"`
	// Record of all returned headers in [key,value,key,value...]
	Headers []string `json:"headers,omitempty"`
	// Multihash of response body (if any)
	Hash string `json:"hash,omitempty"`
}

// WriteSnapshot creates a snapshot record in the DB from a given Url struct
func WriteSnapshot(db sqlQueryExecable, u *Url) error {
	data, err := json.Marshal(u.Headers)
	if err != nil {
		return err
	}
	_, err = db.Exec("insert into snapshots values ($1, $2, $3, $4, $5, $6)", u.Url, u.LastGet.In(time.UTC).Round(time.Second), u.Status, u.DownloadTook, data, u.Hash)
	return err
}

// SnapshotsForUrl returns all snapshots for a given url string
func SnapshotsForUrl(db sqlQueryable, url string) ([]*Snapshot, error) {
	res, err := db.Query("select url, created, status, duration, hash, headers from snapshots where url = $1", url)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	snapshots := make([]*Snapshot, 0)
	for res.Next() {
		c := &Snapshot{}
		if err := c.UnmarshalSQL(res); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, c)
	}

	return snapshots, nil
}

// UnmarshalSQL reads an SQL result into the snapshot receiver
func (s *Snapshot) UnmarshalSQL(row sqlScannable) error {
	var (
		url, hash  string
		created    time.Time
		duration   int64
		status     int
		headerData []byte
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

	*s = Snapshot{
		Url:      url,
		Created:  created.In(time.UTC),
		Status:   status,
		Duration: duration,
		Hash:     hash,
		Headers:  headers,
	}

	return nil
}
