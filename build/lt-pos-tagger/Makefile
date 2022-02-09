-include ../../version
#####################################################################################
service=airenas/tagger
version=$(tagger_version)
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild:
	cd ../.. && docker build -t $(tag) --build-arg BUILD_VERSION=$(version) -f build/lt-pos-tagger/Dockerfile .

dpush: dbuild
	docker push $(tag)
#####################################################################################
.PHONY: dbuild dpush
