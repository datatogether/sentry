# Sentry

<!-- Repo Badges for: Github Project, Slack, License-->

[![GitHub](https://img.shields.io/badge/project-Data_Together-487b57.svg?style=flat-square)](http://github.com/datatogether)
[![Slack](https://img.shields.io/badge/slack-Archivers-b44e88.svg?style=flat-square)](https://archivers-slack.herokuapp.com/)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg)](./LICENSE) 
[![Codecov](https://img.shields.io/codecov/c/github/datatogether/sentry.svg?style=flat-square)](https://codecov.io/gh/datatogether/sentry)

Sentry is a parallelized web crawler written in [Go](https://golang.org) that writes urls, links, & response headers to a Postgres database, then stores the response itself on amazon S3. It keeps a list of “sources”, which use simple string comparison to keep it from wandering outside of designated domains or url paths.

The big difference from other crawlers is a tunable “stale duration”, which will tell the crawler to capture an updated snapshot of the page if the time since the last GET request is older than the stale duration. This gives it a continual “watching” property.

Sentry holds a separate stream of scraping for any url that looks like a dataset. So when it encounters urls that look like `https://foo.com/file.csv`, it assumes that file ending may be a static asset, and places that url on a separate thread for archiving.

## License & Copyright

[Modelled on [project guidelines template](https://github.com/datatogether/roadmap/blob/master/PROJECT.md#license--copyright-readme-block) ]

## Getting Involved

We would love involvement from more people! If you notice any errors or would like to submit changes, please see our [Contributing Guidelines](./.github/CONTRIBUTING.md). 

We use GitHub issues for [tracking bugs and feature requests](https://github.com/datatogether/REPONAME/issues) and Pull Requests (PRs) for [submitting changes](https://github.com/datatogether/REPONAME/pulls)

## ...

## [Optional section(s) on Installation (actually using the service!), Architecture, Dependencies, and Other Considerations]

[fill  out this section if the repo contains deployable/installable code]

## Development

[Step-by-step instructions about how to set up a local dev environment and any dependencies]

## Deployment

[Optional section with deployment instructions]

## Related Projects

In parallel to building this tool, we have engaged in efforts to map the landscape of similar projects:

:eyes: See: [**Comparison of web archiving software**](https://github.com/datatogether/research/tree/master/web_archiving)
