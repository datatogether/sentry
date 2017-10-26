package core

import (
	"time"
)

// Meta is a struct for sharing our knowledge of a url with other services
type Meta struct {
	Url           string            `json:"url"`
	Date          *time.Time        `json:"date,omitempty"`
	HeadersTook   int               `json:"headersTook,omitempty"`
	Id            string            `json:"id"`
	Status        int               `json:"status"`
	ContentSniff  string            `json:"contentSniff,omitempty"`
	RawHeaders    []string          `json:"rawHeaders""`
	Headers       map[string]string `json:"headers"`
	DownloadTook  int               `json:"downloadTook,omitempty"`
	Sha256        string            `json:"sha256"`
	Multihash     string            `json:"multihash"`
	Consensus     *Consensus        `json:"consensus"`
	InboundLinks  []string          `json:"inboundLinks,omitempty"`
	OutboundLinks []string          `json:"outboundLinks,omitempty"`
}
