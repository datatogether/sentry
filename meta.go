package main

import (
	"time"
)

type BaseMeta struct {
	Url          string            `json:"url"`
	Date         time.Time         `json:"date"`
	HeadersTook  time.Duration     `json:"headersTook"`
	Id           string            `json:"id"`
	Status       string            `json:"status"`
	RawHeaders   []string          `json:"rawHeaders""`
	Headers      map[string]string `json:"headers"`
	DownloadTook time.Duration     `json:"downloadTook"`
	File         string            `json:"file"`
}
