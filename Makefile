SHELL=/bin/bash

export GOPATH := $(CURDIR)

.PHONY: all js go

all: js go

go:
	go install -v services/...

js: public/app.js public/lib.js

public/app.js: $(shell find client -name "*.js")
	node_modules/browserify/bin/cmd.js  \
		-t [ babelify --presets [ react ] ] \
		-x d3 \
		-x material-ui/lib/select-field \
		-x material-ui/lib/menus/menu-item \
		-x material-ui/lib/tabs/tabs \
		-x material-ui/lib/tabs/tab \
		-x react \
		-x react-dom \
		-x react-router \
		-x react-tap-event-plugin \
		client/main.js > $@

# node_modules/rebuild is touched by node_modules/.hooks/postinstall everytime `npm install` runs.
# This improves performance of this rule since make only compares one mtime.
public/lib.js: node_modules/rebuild node_modules/.hooks/postinstall
	node_modules/browserify/bin/cmd.js \
		-r d3 \
		-r material-ui/lib/select-field \
		-r material-ui/lib/menus/menu-item \
		-r material-ui/lib/tabs/tabs \
		-r material-ui/lib/tabs/tab \
		-r react \
		-r react-dom \
		-r react-router \
		-r react-tap-event-plugin \
		-o $@

node_modules/.hooks/postinstall:
	mkdir -p $(@D)
	echo touch $(CURDIR)/node_modules/rebuild > $@
	chmod +x $@
