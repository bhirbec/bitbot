SHELL=/bin/bash

export GOPATH := $(CURDIR)

.PHONY: all js go test

all: go js

go:
	go install -v services/...

js:
	cd client && ./node_modules/webpack/bin/webpack.js

test: go
	go test services/...
