compose = docker-compose -f docker-compose.yml
compose-dev = docker-compose -f docker-compose.yml \
	-f docker-compose.override.yml
compose-prod = docker-compose -f docker-compose.yml \
	-f docker-compose.production.yml

COMMIT := $(shell git rev-parse --short HEAD)

update_app_version:
	sed -i .bak 's/APP_VERSION=.*$$/APP_VERSION=$(COMMIT)/' .env

js:
	$(compose-dev) run js-dev ./scripts/minify_js_files.sh

sass:
	$(compose-dev) run sass-dev npm run sass

postcss:
	$(compose-dev) run postcss-dev npm run postcss

assets: sass postcss js

build: update_app_version assets
	$(compose) build

push:
	$(compose) push

check-env:
ifndef DOCKER_MACHINE_NAME
	$(error Please set DOCKER_MACHINE_NAME before deploying)
endif

deploy: check-env build push
	eval $$(docker-machine env ${DOCKER_MACHINE_NAME}) && \
	$(compose-prod) pull && \
	$(compose-prod) up -d

cache_clear:
	eval $$(docker-machine env ${DOCKER_MACHINE_NAME}) && \
	$(compose-prod) exec redis redis-cli flushall

.PHONY: sass postcss assets
