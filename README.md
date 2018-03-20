# Sentry

[![GitHub](https://img.shields.io/badge/project-Data_Together-487b57.svg?style=flat-square)](http://github.com/datatogether)
[![Slack](https://img.shields.io/badge/slack-Archivers-b44e88.svg?style=flat-square)](https://archivers-slack.herokuapp.com/)
[![License](https://img.shields.io/github/license/datatogether/sentry.svg?style=flat-square)](./LICENSE)
[![Codecov](https://img.shields.io/codecov/c/github/datatogether/sentry.svg?style=flat-square)](https://codecov.io/gh/datatogether/sentry)
[![CI](https://img.shields.io/circleci/project/github/datatogether/sentry.svg?style=flat-square)](https://circleci.com/gh/datatogether/sentry)

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

## Usage
Though it has mostly been used with the [Data Together **webapp**](https://github.com/datatogether/webapp), 
**sentry** is a stand-alone web crawler and can be used on its own. It 
currently requires a somewhat elaborate infrastructure and, for instance, it 
can not simply be fed a job over the command line. 

At present, sentry reads crawling instructions directly from a Postgres 
database (see [the schema file](./sql/schema.sql) for details of the 
database structure), and places crawled resources in an S3 bucket. For 
every domain to be crawled, create a record in the `sources` table with 
`crawl` set to true. Sentry will crawl that domain repeatedly. Resources 
will be hashed and stored on S3, where they can be retrieved by 
[**content**](https://github.com/datatogether/content) or any other service 
capable of reverse-engineering the identifying hash. **Other storage backends 
are planned** (see [roadmap](#roadmap), below), and if you are interested in 
helping to develop them please contact us!

## Installation and Configuration

### Docker installation

To get started developing using [Docker](https://store.docker.com/search?type=edition&offering=community) and [Docker Compose](https://docs.docker.com/compose/install/), run:

```shell
$ git clone git@github.com:datatogether/webapp.git
$ cd webapp
$ docker-compose up
```

### Manual installation

1. Install [Go language](https://golang.org/doc/install)
1. Download and build repository
    ```sh
    export GOPATH=$(go env GOPATH)
    mkdir -pv $GOPATH
    cd $GOPATH

    git clone https://github.com/archivers-space/sentry
    cd sentry
    go install
    ```
1. Configure Postgres server and then set connection URL
   ```
   export POSTGRES_DB_URL=postgres://[USERNAME_HERE]:[PASSWORD_HERE]@localhost:[PORT]/[DB_NAME]
   ```
1. Run sentry
    ```sh
    $GOPATH/bin/sentry
    ```
1. Configure S3 buckets [TODO]
    - on production
    - on development (how do you work with them in development env?)

## Roadmap

Two major changes to **sentry** will make it much more generally usable:
- we plan to **shift the storage backend from S3 to IPFS**. Once this is 
accomplished, any local or remote IPFS node can be used as a storage node.
- we are considering **additional mechanisms for adidng crawls to sentry's 
queue**. This should make sentry distinctly more flexible. 

## Related Projects

In parallel to building this tool, we have engaged in efforts to map the
landscape of similar projects:

:eyes: See: [**Comparison of web archiving software**](https://github.com/datatogether/research/tree/master/web_archiving)
