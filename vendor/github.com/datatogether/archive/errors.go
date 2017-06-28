package archive

import "fmt"

var (
	ErrNotFound        = fmt.Errorf("Not Found")
	ErrInvalidResponse = fmt.Errorf("Datastore returned an invalid response")
)
