/*

Sentry is a service for archiving URL's with a clean, current, auditable trail of digital provenance.
It's main job is to issue GET requests to a given URL, record all information related to the request
namely: timestamp, HTTP response headers, sha256 of response body, content-type sniff, content size
and download time. If desirable, sentry can also store the response, currently to amazon S3.

Sentry archives urls in two ways:
	* using a configurable built-in web crawler

*/
package main
