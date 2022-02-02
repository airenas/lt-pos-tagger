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
	go test -v -race ./...
.PHONY: test/unit
#####################################################################################
## code vet and lint
test/lint: 
	go vet ./...
	go get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
.PHONY: test/lint
#####################################################################################
## build docker image
build/lt-pos-tagger:
	cd build/lt-pos-tagger && $(MAKE) clean dpush
.PHONY: build/lt-pos-tagger
#####################################################################################
## build and push lt-pos-tagger docker image
docker/push:
	cd build/lt-pos-tagger && $(MAKE) clean dpush
.PHONY: docker/push
#####################################################################################
## cleans all temporary data
clean:
	cd build/lt-pos-tagger & $(MAKE) clean
