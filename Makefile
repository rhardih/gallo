ifndef DOCKER_CONTEXT
$(error DOCKER_CONTEXT is not set)
endif

COMMIT := $(shell git rev-parse --short HEAD)

# redis container id
rcid := $(shell docker ps -q -f name=gallo-app_redis.1)

update_app_version:
	sed -i .bak 's/APP_VERSION=.*$$/APP_VERSION=$(COMMIT)/' .env

assets: sass postcss js

deploy: build push
	docker context use ${DOCKER_CONTEXT}
	docker compose pull
	docker stack deploy -c docker-compose.yml gallo-app

build: update_app_version
	# We want to build locally
	docker context use default
	docker compose run sass-dev npm run sass
	docker compose run postcss-dev npm run postcss
	docker compose run js-dev ./scripts/minify_js_files.sh
	docker compose -f docker-compose.yml build

push:
	# We want to push from local as well
	docker context use default
	docker compose -f docker-compose.yml push

# Figure out how to do this 
cache_clear:
	docker context use ${DOCKER_CONTEXT}
	docker exec $(rcid) redis-cli flushall

.PHONY: sass postcss assets

