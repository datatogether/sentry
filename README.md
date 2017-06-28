# Sentry
### It watches stuff

Sentry is a parallelized web crawler written in [Go](https://golang.org) that writes urls, links, & response headers to a Postgres database, then stores the response itself on amazon S3. It keeps a list of “sources”, which use simple string comparison to keep it from wandering outside of designated domains or url paths.

The big difference from other crawlers is a tunable “stale duration”, which will tell the crawler to capture an updated snapshot of the page if the time since the last GET request is older than the stale duration. This gives it a continual “watching” property.

Sentry holds a separate stream of scraping for any url that looks like a dataset. So when it encounters urls that look like `https://foo.com/file.csv`, it assumes that file ending may be a static asset, and places that url on a separate thread for archiving.

# Related Projects

In parallel to building this tool, we have engaged in efforts to map the landscape of similar projects:

:eyes: See: [**Comparison of web archiving software**](https://github.com/datatogether/research/tree/master/web_archiving)
