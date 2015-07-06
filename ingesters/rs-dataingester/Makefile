# Use ':=' instead of '=' to avoid multiple evaluation of NOW.
# # Substitute problematic characters with underscore using tr,
# #   make doesn't like spaces and ':' in filenames.
#NOW := $(shell date +"%c" | tr ' :' '__')
NOW := $(shell date +"%s" )

rs-dataingester:
	@echo Pulling dependencies
	@go get ./...
	@echo building rs-dataingester binary
	@go vet
	@go build

docker-image: rs-dataingester
	@echo "Hope you have docker setup and have access ;)"
	docker build -t rs-dataingester .

docker-nolocalgo:
	@echo using centurylink/golang-builder to build docker container
	docker pull centurylink/golang-builder 
	docker run --rm -v ${PWD}:/src -v /var/run/docker.sock:/var/run/docker.sock  centurylink/golang-builder


.PHONY: clean
clean:
	@rm -f rs-dataingester
