package main

import (
	"time"
)

type Meta struct {
	Url          string            `json:"url"`
	Date         time.Time         `json:"date"`
	HeadersTook  int               `json:"headersTook,omitempty"`
	Id           string            `json:"id"`
	Status       int               `json:"status"`
	RawHeaders   []string          `json:"rawHeaders""`
	Headers      map[string]string `json:"headers"`
	DownloadTook int               `json:"downloadTook,omitempty"`
	File         string            `json:"file"`
}
