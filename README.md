# Gallo

## Gallery for Trello

A digital photo frame web application, mainly targeted to run on a [first
generation iPad](https://en.wikipedia.org/wiki/IPad_(1st_generation)). Built on
top of [Trello](https://trello.com).

- Running version Available at [gallo.app](https://gallo.app), although no
  promises on uptime guarantee. Can easily be self-hosted.
- [Changelog](CHANGELOG.md)

### Tech

The application is a *plain old website*, served by a backend written in
[Go](https://golang.org) and sprinkled with a dash of vanilla
[Javascript](https://developer.mozilla.org/en-US/docs/Web/JavaScript).

The Javascript dialect is intentionally older due to targeting [Safari on iOS
5.1.1](https://en.wikipedia.org/wiki/Safari_version_history#Safari_5_3).

Stylesheets are [scss](https://sass-lang.com/).

### External dependencies

- API communication with Trello via
  [adlio/trello](https://github.com/adlio/trello).
- Responsive image polyfill,
  [picturefill](https://scottjehl.github.io/picturefill), and
  [javascript-state-machine](https://github.com/jakesgordon/javascript-state-machine)
  on the Javascript side of things.
- [Pure.css](https://purecss.io) for styling primitives.

## Running

The application is dockerized and includes files to run via [docker
compose](https://docs.docker.com/compose/):

```bash
$ docker-compose up app
```

Apart from the main application service, [redis](https://redis.io/) is included
for caching HTML template renderings and API responses from Trello.

### Environment

The included stubbed [.env](./.env) file lists a number of required and optional
environment variables.

#### Required

- `APP_VERSION` is automatically updated in the file, everytime the
  `update_app_version` make target is run. Will contain the sha of the latest
  commit. The value of this ends up in the **version** attribute of the `<html>`
  tag.
- `HOST` is used to infer the `return_uri` for the Trello authentication flow.
- `REDIS_ADDR` is the ip or hostname of the accompanying Redis server.
- `SESSION_AUTH_KEY` and `SESSION_ENC_KEY` are both 32 character key strings,
  used for session encryption. A tiny Go program for generating these at random
  are available in [keygen.go](./scripts/keygen.go). Run with: `go run
  scripts/keygen.go`.
- `TRELLO_KEY` is a Trello Developer API key. [Get one
  here](https://trello.com/app-key).

#### Optional

These are specifically related to the way the application is running on
[gallo.app](https://gallo.app) and are only relevant if the application is
deployed in a similar setup.

#### Others

- `APP_ENV` and `APP_PATH` are set in the relevant compose files.

### Development

For differences in local development, see
[docker-compose.override.yml](./docker-compose.override.yml) and
[development.Dockerfile](./development.Dockerfile).

Notably, two additional services are run with a *live* file watcher, which
handles conversion of sass to proper css.

## Testing

Assuming you've got the app container running:

```bash
$ docker-compose up -d app
```

Run the Go test suite like so:

```bash
$ docker-compose exec app go test -v ./...
```

## Deploying

The included [Makefile](./Makefile) has a *deploy* target included, which should
make a rudimentary deployment easy:

```bash
$ make deploy
```

This make target will take care of building the application and static assets,
creating the final docker image, pushing it and running the container on your
[DOCKER_HOST](https://docs.docker.com/engine/reference/commandline/cli/#environment-variables).

For a simple single host setup, using [Docker
Machine](https://docs.docker.com/machine/concepts/) with the
[Generic](https://docs.docker.com/machine/drivers/generic/) driver is probably
the easiest way to go.

## Docs

### JSDoc

Comments in Javascript sources has been written in JSDoc format and can be
generated with:

```bash
$ npm run jsdoc
```

Generated docs will be output in `./jsdoc`.
