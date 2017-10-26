package core

import (
	"net/url"

	"github.com/PuerkitoBio/purell"
)

var normalizeFlags = purell.FlagsUsuallySafeGreedy |
	purell.FlagRemoveDuplicateSlashes |
	purell.FlagRemoveFragment |
	purell.FlagLowercaseScheme |
	purell.FlagLowercaseHost |
	purell.FlagUppercaseEscapes

// NormalizeURLString removes inconsitencincies from a given url string
func NormalizeURLString(url string) (string, error) {
	return purell.NormalizeURLString(url, normalizeFlags)
}

// NormalizeURL removes inconsitencincies from a given url
func NormalizeURL(u *url.URL) *url.URL {
	u, err := url.Parse(purell.NormalizeURL(u, normalizeFlags))
	if err != nil {
		// should never happen
		panic(err)
	}
	return u
}
