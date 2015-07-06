# Use ':=' instead of '=' to avoid multiple evaluation of NOW.
# # Substitute problematic characters with underscore using tr,
# #   make doesn't like spaces and ':' in filenames.
#NOW := $(shell date +"%c" | tr ' :' '__')
NOW := $(shell date +"%s" )

rs-platformingester:
	@echo Pulling dependencies
	@go get ./...
	@echo building rs-platformingester binary
	@go vet
	@go build

docker-image: rs-platformingester
	@echo "Hope you have docker setup and have access ;)"
	docker build -t rs-ingester .

docker-nolocalgo:
	@echo using centurylink/golang-builder to build docker container
	docker pull centurylink/golang-builder 
	docker run --rm -v ${PWD}:/src -v /var/run/docker.sock:/var/run/docker.sock  centurylink/golang-builder


.PHONY: clean
clean:
	@rm -f rs-platformingester
