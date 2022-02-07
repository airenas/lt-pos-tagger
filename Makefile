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
.PHONY: test/lint
## run load tests - start services, do load tests, clean services
test/load: 
	cd testing/load && $(MAKE) start all clean	
.PHONY: test/load
## run integration tests - start services, do tests, clean services
test/integration: 
	cd testing/integration && $(MAKE) start test/integration clean	
.PHONY: test/integration
#####################################################################################
## build docker image
build/lt-pos-tagger:
	cd build/lt-pos-tagger && $(MAKE) clean dbuild
.PHONY: build/lt-pos-tagger
#####################################################################################
## build and push lt-pos-tagger docker image
docker/push:
	cd build/lt-pos-tagger && $(MAKE) clean dpush
.PHONY: docker/push
#####################################################################################
## cleans all temporary data
clean:
	cd build/lt-pos-tagger && $(MAKE) clean
.PHONY: clean	
