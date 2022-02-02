-include ../../version
#####################################################################################
DOCKER_IP=$(shell echo $$(ip addr show docker0 | grep -Po 'inet \K[\d.]+'))
TAGGER_VERSION=$(tagger_version)
PORT=8092
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' | sort
.PHONY: help
#####################################################################################
## start all containers for testing, waits for semantika to initialize
start: 
	docker-compose up -d --build
	@echo "wait for semantika ready"
	@docker-compose exec tagger timeout 20 sh -c 'until nc -z semantika 8090; do sleep 1; printf "."; done'
	@echo "\nReady to go"

.PHONY: start
## clean everything, stops docker containers and removes them
clean:
	docker-compose stop && docker-compose rm -f
.PHONY: clean	
#####################################################################################
vu?=10
req?=1000

info:
	@echo IP: $(DOCKER_IP)
## do load test
load: info
	docker run --add-host=host.docker.internal:$(DOCKER_IP) \
		--env EXPECTED_REQ=$(req) \
	    -i loadimpact/k6 run --vus $(vu) --iterations $(req) - <load_simple.js

all: load 
.PHONY: load all 
.EXPORT_ALL_VARIABLES:	