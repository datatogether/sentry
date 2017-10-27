# Sentry

[![GitHub](https://img.shields.io/badge/project-Data_Together-487b57.svg?style=flat-square)](http://github.com/datatogether)
[![Slack](https://img.shields.io/badge/slack-Archivers-b44e88.svg?style=flat-square)](https://archivers-slack.herokuapp.com/)
[![License](https://img.shields.io/github/license/datatogether/sentry.svg?style=flat-square)](./LICENSE)
[![Codecov](https://img.shields.io/codecov/c/github/datatogether/sentry.svg?style=flat-square)](https://codecov.io/gh/datatogether/sentry)
![CI](https://img.shields.io/circleci/project/github/datatogether/sentry.svg?style=flat-square)

Sentry is a parallelized web crawler written in [Go](https://golang.org) that
writes urls, links, & response headers to a Postgres database, then stores the
response itself on amazon S3. It keeps a list of “sources”, which use simple
string comparison to keep it from wandering outside of designated domains or url
paths.

The big difference from other crawlers is a tunable “stale duration”, which will
tell the crawler to capture an updated snapshot of the page if the time since
the last GET request is older than the stale duration. This gives it a continual
“watching” property.

Sentry holds a separate stream of scraping for any url that looks like a
dataset. So when it encounters urls that look like `https://foo.com/file.csv`,
it assumes that file ending may be a static asset, and places that url on a
separate thread for archiving.

## License & Copyright

Copyright (C) 2017 Data Together  
This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, version 3.0.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.

## Getting Involved

We would love involvement from more people! If you notice any errors or would
like to submit changes, please see our
[Contributing Guidelines](./.github/CONTRIBUTING.md).

We use GitHub issues for [tracking bugs and feature requests](https://github.com/datatogether/sentry/issues)
and Pull Requests (PRs) for [submitting changes](https://github.com/datatogether/sentry/pulls)

## Installation
### Docker installation
```
docker compose up
```
### Manual installation
1. Install [Go language](https://golang.org/doc/install)
2. Download and build repository
    ```sh
    export GOPATH=$(go env GOPATH)
    mkdir -pv $GOPATH
    cd $GOPATH

    git clone https://github.com/archivers-space/sentry
    cd sentry
    go install
    ```
3. Configure Postgres server and then set connection URL
   ```
   export POSTGRES_DB_URL=postgres://[USERNAME_HERE]:[PASSWORD_HERE]@localhost:[PORT]/[DB_NAME]
   ```
3. Run sentry
    ```sh
    $GOPATH/bin/sentry
    ```
4. Configure S3 buckets [TODO]
- on production
- on development (how do you work with them in development env?)

## Development

Coming soon!

## Related Projects

In parallel to building this tool, we have engaged in efforts to map the
landscape of similar projects:

:eyes: See: [**Comparison of web archiving software**](https://github.com/datatogether/research/tree/master/web_archiving)
