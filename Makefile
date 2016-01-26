SHELL=/bin/bash

export GOPATH := $(CURDIR)

.PHONY: all js go

all: js go

go:
	go install -v bitbot/cmd/...

js: public/app.js

public/app.js: $(shell find client -name "*.js")
	mkdir -p ./public
	./node_modules/.bin/browserify -t [ babelify --presets [ react ] ] client/main.js > $@
