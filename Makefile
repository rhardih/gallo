ifndef DOCKER_CONTEXT
$(error DOCKER_CONTEXT is not set)
endif

COMMIT := $(shell git rev-parse --short HEAD)

# redis container id
rcid := $(shell docker ps -q -f name=gallo-app_redis.1)

update_app_version:
	sed -i .bak 's/APP_VERSION=.*$$/APP_VERSION=$(COMMIT)/' .env

deploy: build push # deprecated - moved to k3s
	docker -c ${DOCKER_CONTEXT} compose pull
	docker -c ${DOCKER_CONTEXT} stack deploy -c docker-compose.yml gallo-app

build: update_app_version
	docker -c default compose run sass-dev npm run sass
	docker -c default compose run postcss-dev npm run postcss
	docker -c default compose run js-dev ./scripts/minify_js_files.sh
	docker -c default compose -f docker-compose.yml build

push:
	docker -c default compose -f docker-compose.yml push

# Figure out how to do this 
cache_clear:
	docker -c ${DOCKER_CONTEXT} exec $(rcid) redis-cli flushall

.PHONY: sass postcss assets

