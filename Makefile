CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src/github.com/whosonfirst/go-whosonfirst-rasterzen; then rm -rf src/github.com/whosonfirst/go-whosonfirst-rasterzen; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-rasterzen
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-index"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-geojson-v2"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-rasterzen"
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/go-spatial src/github.com/
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/jtacoma src/github.com/
	mv src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/tidwall src/github.com/
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/aws src/github.com/
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/whosonfirst/go-whosonfirst-cli src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/whosonfirst/go-whosonfirst-aws src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/whosonfirst/go-whosonfirst-log src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/whosonfirst/go-whosonfirst-cache src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-rasterzen/vendor/github.com/whosonfirst/go-whosonfirst-cache-s3 src/github.com/whosonfirst/

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	# go fmt *.go
	go fmt cmd/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/wof-rasterzen-seed cmd/wof-rasterzen-seed.go

