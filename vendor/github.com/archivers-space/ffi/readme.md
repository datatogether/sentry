# ffi
--
    import "github.com/archivers-space/ffi"

FFI - File Format identification is a package for making sensible guesses about
file formats from Url strings. FFI is structured as a package to clear the way
for future improvements.

The biggest avenue for improvement may be through integration with PRONOM:
http://www.nationalarchives.gov.uk/PRONOM/Default.aspx#

PRONOM is a repository of file format signatures stored as XML. Currently the
best tool for working with PRONOM is DROID, which is ill-suited to being run in
web service oriented architectures.

As a suggested path for implementation, one could first focus on wrapping DROID
in a service that invokes droid-cli with a given set of bytes. This would happen
in a separate pacakge from this one. Later we could sit down & learn how to
properly parse PRONOM xml signature files & do byte comparison, which could bue
implemented in this package. Anyone investigating this avenue may wish to also
review some of the work accomplished translating the PRONOM registry:
http://the-fr.org

UPDATE 2017-04-24: looks like the File Information Tool Set from harvard may in
fact be a better place to start: http://projects.iq.harvard.edu/fits/home

## Usage

#### func  ExtensionMimeType

```go
func ExtensionMimeType(extenstion string) (string, error)
```
ExtensionMimeType returns a MIME-Type for a given file extension

#### func  FilenameFromUrlString

```go
func FilenameFromUrlString(rawUrl string) (string, error)
```
FilenameFromUrlString returns a file with extension if the url looks like it
resolves to a filename, otherwise it returns an empty string

#### func  FilenameMimeType

```go
func FilenameMimeType(filename string) (string, error)
```
FilenameMimeType gives the mimetype of filename

#### func  MimeType

```go
func MimeType(io.Reader) (string, error)
```
MimeType from an io.Reader is the magic we seek

#### func  MimeTypeExtension

```go
func MimeTypeExtension(mimeType string) (string, error)
```
MimeTypeExtension returns an extension for a given MIME-type

#### func  SetExtension

```go
func SetExtension(filename, mimeType string) (string, error)
```
SetExtension strips any current extension from filename, replacing it with
mimeType's corresponding extension
