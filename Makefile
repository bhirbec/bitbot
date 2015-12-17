SHELL=/bin/bash

export GOPATH := $(CURDIR)

.PHONY: all js go

all: js go

go:
	go install -v bitbot/cmd/...

js: public/build/main.js

public/build/main.js: $(shell find client -name "*.js" -o -name "*.jsx")
	mkdir -p ./public/build
	./node_modules/.bin/browserify -t reactify client/main.js > $@
