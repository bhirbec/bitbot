SHELL=/bin/bash

export GOPATH := $(CURDIR)

.PHONY: all js go

all: js go

go:
	go install -v bitbot/cmd/...

js: public/app.js

public/app.js: $(shell find client -name "*.js" -o -name "*.jsx")
	mkdir -p ./public
	./node_modules/.bin/browserify -t reactify client/main.js > $@
