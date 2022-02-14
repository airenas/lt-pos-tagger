-include Makefile.options
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' | sort 
.PHONY: help
#####################################################################################
## call units tests
test/unit: 
	go test -v -race -count=1 ./...
.PHONY: test/unit
#####################################################################################
## code vet and lint
test/lint: 
	go vet ./...
	go get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
	go mod tidy
.PHONY: test/lint
## run load tests - start services, do load tests, clean services
test/load:
	cd testing/load && $(MAKE) start all clean || ( $(MAKE) clean; exit 1; ) 
.PHONY: test/load
## run integration tests - start services, do tests, clean services
test/integration:
	cd testing/integration && $(MAKE) start test/integration clean || ( $(MAKE) clean; exit 1; ) 	
.PHONY: test/integration
#####################################################################################
## build docker image
docker/build:
	cd build/lt-pos-tagger && $(MAKE) dbuild
.PHONY: docker/build
#####################################################################################
## build and push lt-pos-tagger docker image
docker/push:
	cd build/lt-pos-tagger && $(MAKE) dpush
.PHONY: docker/push
#####################################################################################
## cleans all temporary data
clean:
	go clean
	go mod tidy
	cd testing/integration && $(MAKE) clean
	cd testing/load && $(MAKE) clean
.PHONY: clean	
