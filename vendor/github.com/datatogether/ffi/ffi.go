/*
FFI - File Format identification is a package for making sensible
guesses about file formats from Url strings. FFI is structured as a package
to clear the way for future improvements.

The biggest avenue for improvement may be through integration with PRONOM:
http://www.nationalarchives.gov.uk/PRONOM/Default.aspx#

PRONOM is a repository of file format signatures stored as XML.
Currently the best tool for working with PRONOM is DROID, which is ill-suited to
being run in web service oriented architectures.

As a suggested path for implementation, one could first focus on wrapping DROID
in a service that invokes droid-cli with a given set of bytes. This would happen
in a separate pacakge from this one. Later we could sit down & learn how to
properly parse PRONOM xml signature files & do byte comparison, which could
bue implemented in this package. Anyone investigating this avenue may wish to also
review some of the work accomplished translating the PRONOM registry:
http://the-fr.org

UPDATE 2017-04-24: looks like the File Information Tool Set from harvard may in fact be a better
place to start: http://projects.iq.harvard.edu/fits/home

*/
package ffi

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
)

// FilenameFromUrlString returns a file with extension if the url
// looks like it resolves to a filename, otherwise it returns an
// empty string
func FilenameFromUrlString(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}
	return filepath.Base(u.Path), nil
}

// FilenameMimeType gives the mimetype of filename
func FilenameMimeType(filename string) (string, error) {
	return MimeTypeExtension(filepath.Ext(filename))
}

// MimeTypeExtension returns an extension for a given MIME-type
func MimeTypeExtension(mimeType string) (string, error) {
	ext := mappings[mimeType]
	if ext == "" {
		return "", fmt.Errorf("unrecognized MIME-Type: '%s'", mimeType)
	}
	return ext, nil
}

// ExtensionMimeType returns a MIME-Type for a given file extension
func ExtensionMimeType(extenstion string) (string, error) {
	mt := mappings[extenstion]
	if mt == "" {
		return "", fmt.Errorf("unrecognized extension: '%s'", extenstion)
	}
	return mt, nil
}

// SetExtension strips any current extension from filename,
// replacing it with mimeType's corresponding extension
func SetExtension(filename, mimeType string) (string, error) {
	suffix := filepath.Ext(filename)
	ext, err := MimeTypeExtension(mimeType)
	if err != nil {
		return filename, err
	}
	return filename[:len(filename)-len(suffix)] + ext, nil
}

// MimeType from an io.Reader is the magic we seek
func MimeType(io.Reader) (string, error) {
	return "", fmt.Errorf("we don't yet support MimeType from an io.Reader, maybe you can be the one to write this feature?")
}
