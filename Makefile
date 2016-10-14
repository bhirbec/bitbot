SHELL=/bin/bash

export GOPATH := $(CURDIR)

.PHONY: all js go

all: js go

go:
	go install -v services/...

js: public/app.js public/lib.js

public/app.js: $(shell find client -name "*.js")
	node_modules/browserify/bin/cmd.js  \
		-t [ babelify --presets [ es2015 react ] ] \
		-x d3 \
		-x material-ui/SelectField \
		-x material-ui/MenuItem \
		-x material-ui/Tabs \
		-x material-ui/styles/MuiThemeProvider \
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
		-r material-ui/SelectField \
		-r material-ui/MenuItem \
		-r material-ui/Tabs \
		-r material-ui/styles/MuiThemeProvider \
		-r react \
		-r react-dom \
		-r react-router \
		-r react-tap-event-plugin \
		-o $@

node_modules/.hooks/postinstall:
	mkdir -p $(@D)
	echo touch $(CURDIR)/node_modules/rebuild > $@
	chmod +x $@
