package archive

import (
	"github.com/pborman/uuid"
)

func NewUuid() string {
	return uuid.New()
}
