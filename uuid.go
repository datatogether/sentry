package main

import (
	"github.com/pborman/uuid"
)

func NewUuid() string {
	return uuid.New()
}
